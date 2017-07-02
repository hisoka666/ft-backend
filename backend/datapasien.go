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

// func GetListPasien(c context.Context, email string, m, y int) []Pasien {
// 	// ctx := appengine.NewContext(r)
// 	// email, _, _ := AppCtx(ctx, "", "", "", "")
// 	monIn := DatebyInt(m, y)
// 	q := datastore.NewQuery("KunjunganPasien").Filter("Dokter =", email).Filter("Hide =", false).Order("-JamDatang")
// 	list := IterateList(c, w, q, monIn)
// 	return list
// }

// func IterateList(c context.Context, w http.ResponseWriter, q *datastore.Query, mon time.Time) []ListPasien {
// 	t := q.Run(c)
// 	monAf := mon.AddDate(0, 1, 0)
// 	var daf KunjunganPasien
// 	var tar Pasien
// 	// var pts DataPasien
// 	var list []Pasien
// 	for {
// 		k, err := t.Next(&daf)
// 		if err == datastore.Done {
// 			break
// 		}
// 		if err != nil {
// 			fmt.Fprintln(w, "Error Fetching Data: ", err)
// 		}
// 		daf.JamDatang = daf.JamDatang.Add(time.Duration(8) * time.Hour)
// 		jam := UbahTanggal(daf.JamDatang, daf.ShiftJaga)
// 		if jam.After(monAf) == true {
// 			continue
// 		}
// 		if jam.Before(mon) == true {
// 			break
// 		}
// 		if daf.Hide == true {
// 			continue
// 		}
// 		tar.TanggalFinal = jam.Format("02-01-2006")

// 		nocm := k.Parent()
// 		tar.NomorCM = nocm.StringID()

// 		err = datastore.Get(ctx, nocm, &pts)
// 		if err != nil {
// 			fmt.Fprintln(w, "Error Fetching Data Pasien: ", err)
// 		}

// 		tar.NamaPasien = ProperTitle(pts.NamaPasien)
// 		tar.Diagnosis = ProperTitle(daf.Diagnosis)
// 		tar.ShiftJaga = daf.ShiftJaga
// 		tar.LinkID = k.Encode()

// 		if daf.GolIKI == "1" {
// 			tar.IKI1 = "1"
// 			tar.IKI2 = ""
// 		} else {
// 			tar.IKI1 = ""
// 			tar.IKI2 = "1"
// 		}

// 		list = append(list, tar)
// 	}

// 	return list
// }
