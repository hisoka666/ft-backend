package backend

import (
	"strconv"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

// type WebView struct {
// 	Token    string    `json:"token"`
// 	User     string    `json:"user"`
// 	Pasien   []Pasien  `json:"pasien"`
// 	//IKI      []List    `json:"list"`
// }

// //Ini digunakan untuk view web, IKI1 dan IKI2 harus dipisah
// type Pasien struct {
// 	TglKunjungan string `json:"tgl"`
// 	ShiftJaga    string `json:"shift"`
// 	NoCM         string `json:"nocm"`
// 	NamaPasien   string `json:"nama"`
// 	Diagnosis    string `json:"diag"`
// 	IKI1         string `json:"iki1"`
// 	IKI2         string `json:"iki2"`
// 	LinkID       string `json:"link"`
// }

// //Ini untuk menyimpan jumlah iki yang diperoleh
// type List struct {
// 	TglJaga      string `json:"tgl"`
// 	//ShiftJaga    string `json:"shift"`
// 	SumIKI1         string `json:"iki1"`
// 	SumIKI2         string `json:"iki2"`
// }

// type KunjunganPasien struct {
// 	Diagnosis, LinkID      string
// 	GolIKI, ATS, ShiftJaga string
// 	JamDatang              time.Time
// 	Dokter                 string
// 	Hide                   bool
// 	JamDatangRiil          time.Time
// }

// type DataPasien struct {
// 	NamaPasien, NomorCM, JenKel, Alamat string
// 	TglDaftar, Umur                     time.Time
// }

func GetLast100(c context.Context, email string) []Pasien {
	q := datastore.NewQuery("KunjunganPasien").Limit(100).Filter("Dokter =", email).Order("-JamDatang")
	var m []Pasien

	t := q.Run(c)

	for {
		var j KunjunganPasien
		k, err := t.Next(&j)
		if err == datastore.Done {
			break
		}

		if err != nil {
			LogError(c, err)
			break
		}

		n := ConvertDatastore(c, j, k)

		m = append(m, *n)

	}
	return m

}

func ConvertDatastore(c context.Context, n KunjunganPasien, k *datastore.Key) *Pasien {
	tanggal := UbahTanggal(n.JamDatang, n.ShiftJaga)
	nocm, namapts := GetDataPts(c, k)

	m := &Pasien{
		TglKunjungan: tanggal,
		ShiftJaga:    n.ShiftJaga,
		NoCM:         nocm,
		NamaPasien:   namapts,
		Diagnosis:    n.Diagnosis,
		IKI1:         "",
		IKI2:         "",
		LinkID:       k.Encode(),
	}

	if n.GolIKI == "1" {
		m.IKI1 = "1"
		m.IKI2 = ""
	} else {
		m.IKI1 = ""
		m.IKI2 = "1"
	}

	return m
}

func GetDataPts(c context.Context, k *datastore.Key) (no, nama string) {
	var p DataPasien
	keypar := k.Parent()
	err := datastore.Get(c, keypar, &p)
	if err != nil {
		log.Errorf(c, "error getting data: %v", err)
	}
	no = keypar.StringID()
	nama = p.NamaPasien
	return no, nama
}

func GetNamaByNoCM(c context.Context, nocm string) DataPasien {
	var pts DataPasien
	parKey := datastore.NewKey(c, "IGD", "fasttrack", 0, nil)
	ptsKey := datastore.NewKey(c, "DataPasien", nocm, 0, parKey)

	err := datastore.Get(c, ptsKey, &pts)
	if err != nil && err == datastore.ErrNoSuchEntity {
		return "data-not-available"
	}

	return pts
}

func GetListIKI(pts []Pasien, m, y int) []List {
	for i, j := 0, len(pts)-1; i < j; i, j = i+1, j-1 {
		pts[i], pts[j] = pts[j], pts[i]
	}
	mo := DatebyInt(m, y)
	wkt := time.Date(mo.Year(), mo.Month(), 1, 0, 0, 0, 0, time.UTC)
	strbl := wkt.AddDate(0, 1, -1).Format("2")
	bl, _ := strconv.Atoi(strbl)
	var ikiBulan []List
	ikiBulan = append(ikiBulan, List{})
	for h := 1; h <= bl; h++ {
		dataIKI := List{}
		q := time.Date(mo.Year(), mo.Month(), h, 0, 0, 0, 0, time.UTC).Format("02-01-2006")
		var u1, u2 int
		for _, v := range pts {
			if v.TglKunjungan != q {
				continue
			}
			if v.IKI1 == "1" {
				u1++
			} else {
				u2++
			}
		}

		if u1 == 0 && u2 == 0 {
			continue
		}
		dataIKI.TglJaga = q
		dataIKI.SumIKI1 = string(u1)
		dataIKI.SumIKI2 = string(u2)

		ikiBulan = append(ikiBulan, dataIKI)
	}

	return ikiBulan
}
