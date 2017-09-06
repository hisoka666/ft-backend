package backend

import (
	"strconv"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

func GetListSupbyKursor(c context.Context, tgl string) ([]SupervisorListPasien, time.Time) {
	kurKey, _ := DatastoreKey(c, "KursorIGD", tgl, "", "")
	// log.Infof(c, "Key adalah: %s", kurKey)
	yr, _ := strconv.Atoi(tgl[:4])
	mo, _ := strconv.Atoi(tgl[5:7])
	zone, _ := time.LoadLocation("Asia/Makassar")
	q := datastore.NewQuery("KunjunganPasien").Filter("Hide=", false).Order("-JamDatang")
	monini := time.Date(yr, time.Month(mo), 0, 0, 0, 0, 0, zone)
	// jmlhari := monini.Day()
	kur := KursorIGD{}
	err := datastore.Get(c, kurKey, &kur)
	if err != nil {
		LogError(c, err)
	}
	log.Infof(c, "Point adalah : %v", kur.Point)

	kursor, err := datastore.DecodeCursor(kur.Point)
	if err != nil {
		LogError(c, err)
	}
	log.Infof(c, "Kursor adalah: %v", kursor)

	q = q.Start(kursor)

	var m []SupervisorListPasien

	t := q.Run(c)
	// log.Infof(c, "Sampai di sini berarti start kursor jalan")
	monin := time.Date(yr, time.Month(mo), 1, 8, 0, 0, 0, zone)
	// log.Infof(c, "tanggal satu bulan ini adalah: %v", monin)
	// strmonin := tgl + "/01"
	// monin, err := time.Parse("2006/01/02", strmonin)
	// log.Infof(c, "Tanggal satu adalah : %v", monin)

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
		// log.Infof(c, "Jam datang adalah : %v", j.JamDatang)
		// jamAdjust := AdjustTime(j.JamDatang, j.ShiftJaga)
		// log.Infof(c, "Apakah sebelum? %v", jamAdjust.Before(monin))
		if j.JamDatangRiil.Before(monin) == true {
			break
		}

		n := ConvertData(k, j, zone)
		// log.Infof(c, "Tanggal ini isinya: %v", n)

		m = append(m, *n)
	}

	for i, j := 0, len(m)-1; i < j; i, j = i+1, j-1 {
		m[i], m[j] = m[j], m[i]
	}

	// log.Infof(c, "list adalah : %v", m)
	return m, monini
}
func PerHariPerBulan(c context.Context, n *[]SupervisorListPasien, bul time.Time) ([]int, []Departemen, []PerShift) {
	jml := []int{}
	deptlist := []Departemen{}
	shf := []PerShift{}
	har := bul.Day()
	for i := 1; i <= har; i++ {
		perhari := bul.AddDate(0, 0, i)
		log.Infof(c, "Tanggal: %v", perhari)
		kun := []SupervisorListPasien{}
		jph := 0
		for _, v := range *n {
			if v.TglKunjungan.Before(perhari) {
				continue
			}
			if v.TglKunjungan.After(perhari.AddDate(0, 0, 1)) {
				continue
			}
			kun = append(kun, v)
			jph++
		}
		shf = append(shf, perShift(kun, perhari))
		deptlist = append(deptlist, *perBagian(kun))
		jml = append(jml, jph)
	}

	return jml, deptlist, shf
}

func CreateEndKursor(c context.Context, email string) {

	q := datastore.NewQuery("KunjunganPasien").Filter("Dokter=", email).Filter("Hide=", false).Order("-JamDatang")
	t := q.Run(c)
	kur := Kursor{}
	kun := KunjunganPasien{}
	// days := time.Date(yr,time.Month(mo),0,0,0,0,0,zone).Day()
	// mon := time.Date()
	zone, _ := time.LoadLocation("Asia/Makassar")
	todayIs := time.Now().In(zone)
	hariini := time.Date(todayIs.Year(), todayIs.Month(), 1, 0, 0, 0, 0, zone)
	tgl := hariini.AddDate(0, -1, 0).Format("2006/01")
	_, kurKey := DatastoreKey(c, "Dokter", email, "Kursor", tgl)
	log.Infof(c, "Waktu lokal adalah: %v", hariini)
	for {
		_, err := t.Next(&kun)
		if err == datastore.Done {
			break
		}
		if err != nil {
			LogError(c, err)
		}

		jamEdit := AdjustTime(kun.JamDatang, kun.ShiftJaga)
		log.Infof(c, "Jamedit adalah: %v", jamEdit)
		log.Infof(c, "Apakah hari ini sebelum tanggal 1? %v", jamEdit.Before(hariini))
		if jamEdit.Before(hariini) == true {
			cursor, _ := t.Cursor()
			kur.Point = cursor.String()
			if _, err := datastore.Put(c, kurKey, &kur); err != nil {
				LogError(c, err)
			}
			break
		}
	}
}
func CreateKursorIGD(c context.Context) {
	// c := appengine.NewContext(r)
	q := datastore.NewQuery("KunjunganPasien").Filter("Hide=", false).Order("-JamDatang")
	t := q.Run(c)
	kur := KursorIGD{}
	kun := KunjunganPasien{}
	// days := time.Date(yr,time.Month(mo),0,0,0,0,0,zone).Day()
	// mon := time.Date()
	zone, _ := time.LoadLocation("Asia/Makassar")
	todayIs := time.Now().In(zone)
	hariini := time.Date(todayIs.Year(), todayIs.Month(), 1, 8, 0, 0, 0, zone)
	tgl := hariini.AddDate(0, -1, 0).Format("2006/01")
	// _, kurKey := DatastoreKey(c, "KursorIGD", tgl, "", "")
	// log.Infof(c, "Waktu lokal adalah: %v", hariini)
	// for i := 8; i > 1; i-- {
	// 	zone, _ := time.LoadLocation("Asia/Makassar")
	// 	todayIs := time.Now().In(zone)
	// 	hariini := time.Date(todayIs.Year(), time.Month(i), 1, 8, 0, 0, 0, zone)
	// 	tgl := hariini.Format("2006/01")
	kurKey, _ := DatastoreKey(c, "KursorIGD", tgl, "", "")
	// 	// log.Infof(c, "Key adalah: %v", kurKey)
	for {
		_, err := t.Next(&kun)
		if err == datastore.Done {
			log.Infof(c, "Data tidak ditemukan!")
			break
		}
		if err != nil {
			log.Infof(c, "Kesalahan membaca database: %v", err)
			break
			// LogError(c, err)
		}

		// jamEdit := AdjustTime(kun.JamDatang, kun.ShiftJaga)
		// log.Infof(c, "Jamedit adalah: %v", jamEdit)
		// log.Infof(c, "Apakah hari ini sebelum tanggal 1? %v", jamEdit.Before(hariini))
		if kun.JamDatang.Before(hariini) == true {
			cursor, _ := t.Cursor()
			kur.Point = cursor.String()
			kur.Bulan = tgl
			k, err := datastore.Put(c, kurKey, &kur)
			if err != nil {
				log.Errorf(c, "Kesalahan menulis database: %v", err)
				break
			}
			// log.Infof(c, "key kursor adalah: %v", k)
			// err = datastore.Get(c, k, &kur)
			// if err != nil {
			// 	log.Errorf(c, "Gagal memperoleh data : %v", err)
			// 	break
			// }
			// log.Infof(c, "Kursor adalah: %v", kur.Point)
			// log.Infof(c, "key kursor adalah: %v", k)
			log.Infof(c, "Berhasil menambahkan kursor %v", k)
			break
		}
	}
}
func AdjustTime(t time.Time, s string) time.Time {
	zone, _ := time.LoadLocation("Asia/Makassar")
	idn := t.In(zone)
	jam := idn.Hour()
	// tglstring := ""
	if jam < 12 && s == "3" {
		// tglstring =
		return idn.AddDate(0, 0, -1)
	} else {
		// tglstring =
		return idn
	}
}

func DatastoreKey(ctx context.Context, kind1 string, id1 string, kind2 string, id2 string) (*datastore.Key, *datastore.Key) {
	gpKey := datastore.NewKey(ctx, "IGD", "fasttrack", 0, nil)
	parKey := datastore.NewKey(ctx, kind1, id1, 0, gpKey)
	chldKey := datastore.NewKey(ctx, kind2, id2, 0, parKey)

	return parKey, chldKey
}
