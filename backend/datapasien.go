package backend

import (
	"strconv"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

func GetLast100(c context.Context, email string) []Pasien {
	q := datastore.NewQuery("KunjunganPasien").Filter("Dokter =", email).Filter("Hide =", false).Order("-JamDatang").Limit(100)
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

//todo : ubahTanggalPasien
func ConvertDatastore(c context.Context, n KunjunganPasien, k *datastore.Key) *Pasien {
	tanggal := UbahTanggal(n.JamDatang, n.ShiftJaga)
	nocm, namapts := GetDataPts(c, k)

	m := &Pasien{
		TglKunjungan: tanggal,
		ShiftJaga:    n.ShiftJaga,
		NoCM:         nocm,
		NamaPasien:   namapts,
		Diagnosis:    n.Diagnosis,
		IKI:          "",
		LinkID:       k.Encode(),
	}

	if n.GolIKI == "1" {
		m.IKI = "1"
	} else {
		m.IKI = ""
	}

	return m
}

func GetKunPasien(c context.Context, link string) *Pasien {
	var kun KunjunganPasien
	var dat DataPasien

	log.Infof(c, "Link adalah : %v", link)
	keyKun, err := datastore.DecodeKey(link)
	if err != nil {
		LogError(c, err)
	}
	err = datastore.Get(c, keyKun, &kun)
	if err != nil {
		LogError(c, err)
	}
	log.Infof(c, "Diagnosis adalah: %v", kun.Diagnosis)
	keyPts := keyKun.Parent()
	log.Infof(c, "No Cm adalah: %v", keyPts.StringID())
	err = datastore.Get(c, keyPts, &dat)
	if err != nil {
		LogError(c, err)
	}

	pts := &Pasien{
		NoCM:       keyPts.StringID(),
		NamaPasien: dat.NamaPasien,
		Diagnosis:  kun.Diagnosis,
		ATS:        kun.ATS,
		ShiftJaga:  kun.ShiftJaga,
		Bagian:     kun.Bagian,
		IKI:        kun.GolIKI,
		LinkID:     link,
		TglAsli:    kun.JamDatangRiil.Add(time.Duration(8) * time.Hour),
	}

	return pts
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

func GetNamaByNoCM(c context.Context, nocm string) (DataPasien, error) {
	var pts DataPasien
	parKey := datastore.NewKey(c, "IGD", "fasttrack", 0, nil)
	ptsKey := datastore.NewKey(c, "DataPasien", nocm, 0, parKey)

	err := datastore.Get(c, ptsKey, &pts)
	if err != nil && err != datastore.ErrNoSuchEntity {
		return pts, err
	}

	return pts, nil
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
			if v.IKI == "1" {
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

func UpdateEntri(c context.Context, n *Pasien) (*Pasien, error) {

	kun := &KunjunganPasien{}
	pts := &DataPasien{}

	keyKun, err := datastore.DecodeKey(n.LinkID)
	if err != nil {
		m := &Pasien{
			StatusServer: "kesalahan-decoding-Key",
		}
		return m, err
	}
	keyPts := keyKun.Parent()

	err = datastore.Get(c, keyKun, kun)
	if err != nil {
		m := &Pasien{
			StatusServer: "kesalahan-database-get-kunjungan-pts",
		}
		return m, err
	}

	kun.Diagnosis = n.Diagnosis
	kun.ATS = n.ATS
	kun.GolIKI = n.IKI
	kun.ShiftJaga = n.ShiftJaga
	kun.Bagian = n.Bagian
	n.TglKunjungan = UbahTanggal(kun.JamDatang, kun.ShiftJaga)

	err = datastore.Get(c, keyPts, pts)
	if err != nil {
		m := &Pasien{
			NoCM: "kesalahan-database-get-datapts",
		}
		return m, err
	}
	pts.NamaPasien = n.NamaPasien

	if _, err := datastore.Put(c, keyKun, kun); err != nil {
		m := &Pasien{
			StatusServer: "kesalahan-database-put-kunjungan-failed",
		}
		return m, err
	}

	if _, err := datastore.Put(c, keyPts, pts); err != nil {
		m := &Pasien{
			StatusServer: "kesalahan-database-put-datapts-failed",
		}
		return m, err
	}

	if kun.GolIKI == "1" {
		n.IKI = "1"
	} else {
		n.IKI = ""
	}
	n.NoCM = keyPts.StringID()
	n.StatusServer = "OK"
	log.Infof(c, string(ConvertJSON(n)))
	return n, nil
}

func DeleteEntri(c context.Context, link string) *Pasien {
	kun := &KunjunganPasien{}

	keyKun, err := datastore.DecodeKey(link)
	if err != nil {
		LogError(c, err)
	}
	err = datastore.Get(c, keyKun, kun)
	if err != nil {
		LogError(c, err)
	}

	kun.Hide = true
	if _, err := datastore.Put(c, keyKun, kun); err != nil {
		m := &Pasien{
			StatusServer: "kesalahan-database-put-kunjungan-failed",
		}
		LogError(c, err)
		return m
	}

	m := &Pasien{
		StatusServer: "OK",
	}

	return m
}
