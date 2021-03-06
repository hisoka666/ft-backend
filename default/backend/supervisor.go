package backend

import (
	"fmt"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

func SupervisorPage(c context.Context, token string, user string) *MainView {
	skrng, zone := CreateTime()
	awal := time.Date(skrng.Year(), skrng.Month(), 1, 8, 0, 0, 0, zone)
	// awal = awal.AddDate(0, -1, 0)
	q := datastore.NewQuery("KunjunganPasien").Filter("JamDatang >=", awal).Filter("Hide=", false).Order("-JamDatang")
	t := q.Run(c)
	sup := SupervisorList{}
	kun := []SupervisorListPasien{}
	dat := KunjunganPasien{}
	for {
		k, err := t.Next(&dat)
		if err == datastore.Done {
			log.Infof(c, "Data tidak ditemukan")
			break
		}
		if err != nil {
			LogError(c, err)
		}

		m := ConvertData(k, dat, zone)
		kun = append(kun, m)
	}

	for i, j := 0, len(kun)-1; i < j; i, j = i+1, j-1 {
		kun[i], kun[j] = kun[j], kun[i]
	}
	sup.Token = token
	sup.StatusServer = "OK"
	sup.SupervisorName = user
	// sup.ListPasien = kun
	har, dep, shift := PerHari(c, kun)
	sup.PerHari = har
	sup.PerDeptPerHari = dep
	sup.PerShiftPerHari = shift
	// log.Infof(c, "List adalah: %v", sup.ListPasien)
	main := &MainView{
		Token:      token,
		User:       user,
		Supervisor: sup,
		Peran:      "supervisor",
		Bulan:      getIGDBulan(c),
	}
	log.Infof(c, "List bulan adalah: %v", getIGDBulan(c))
	return main

}
func getIGDBulan(c context.Context) []string {
	kur := []KursorIGD{}
	q := datastore.NewQuery("KursorIGD").Order("-Bulan")
	n, err := q.GetAll(c, &kur)
	if err != nil {
		log.Errorf(c, "Kesalahan membaca kursor database: %v", err)
		return nil
	}
	var list []string
	for _, v := range n {
		list = append(list, v.StringID())
	}

	return list

}
func ConvertData(k *datastore.Key, n KunjunganPasien, zone *time.Location) SupervisorListPasien {
	m := SupervisorListPasien{
		TglKunjungan: n.JamDatangRiil.In(zone),
		ATS:          n.ATS,
		Dept:         n.Bagian,
		Diagnosis:    n.Diagnosis,
		LinkID:       k.Encode(),
	}
	return m
}

func CreateTime() (time.Time, *time.Location) {
	t := time.Now()
	zone, err := time.LoadLocation("Asia/Makassar")
	if err != nil {
		fmt.Println("Err: ", err.Error())
	}
	jam := t.In(zone)
	return jam, zone
}

func perBagian(n []SupervisorListPasien) *Departemen {
	var interna, bedah, anak, obgyn, saraf, anes, psik, tht, kulit, jant, um, mata, mod int
	for _, v := range n {
		switch v.Dept {
		case "1":
			interna++
		case "2":
			bedah++
		case "3":
			anak++
		case "4":
			obgyn++
		case "5":
			saraf++
		case "6":
			anes++
		case "7":
			psik++
		case "8":
			tht++
		case "9":
			kulit++
		case "10":
			jant++
		case "11":
			um++
		case "12":
			mata++
		case "13":
			mod++
		}
	}
	m := &Departemen{
		Interna:   interna,
		Bedah:     bedah,
		Anak:      anak,
		Obgyn:     obgyn,
		Saraf:     saraf,
		Anestesi:  anes,
		Psikiatri: psik,
		THT:       tht,
		Kulit:     kulit,
		Kardio:    jant,
		Umum:      um,
		MOD:       mod,
		Mata:      mata,
	}
	// log.Infof(c, "")
	return m
}

func PerHari(c context.Context, n []SupervisorListPasien) ([]int, []Departemen, []PerShift) {
	t, z := CreateTime()
	// t := d.AddDate(0, -1, 0)
	log.Infof(c, "list untuk perhari adalah: %v", n)
	hari := time.Date(t.Year(), t.Month(), 0, 0, 0, 0, 0, z).Day()
	jml := []int{}
	deptlist := []Departemen{}
	shf := []PerShift{}
	// var jmlperhari *int
	for i := 1; i < hari; i++ {
		perhari := time.Date(t.Year(), t.Month(), i, 8, 0, 0, 0, z)
		// log.Infof(c, "Tanggal : %v", perhari)
		// log.Infof(c, "Tanggal besok : %v", perhari.AddDate(0, 0, 1))
		kun := []SupervisorListPasien{}
		jph := 0
		// dat := &SupervisorListPasien{}
		for _, v := range n {
			// log.Infof(c, "Is before true? %v", v.TglKunjungan.Before(perhari))
			// log.Infof(c, "Is after true? %v", v.TglKunjungan.After(perhari.AddDate(0, 0, 1)))
			// jml = []int{}
			// log.Infof(c, "Tanggal besok : %v", perhari.AddDate(0, 0, 1))
			if v.TglKunjungan.Before(perhari) {
				continue
			}
			if v.TglKunjungan.After(perhari.AddDate(0, 0, 1)) {
				continue
			}
			log.Infof(c, "Data hari ini: %v", v)

			kun = append(kun, v)
			jph++
		}
		shf = append(shf, perShift(kun, perhari))
		deptlist = append(deptlist, *perBagian(kun))
		jml = append(jml, jph)

	}

	return jml, deptlist, shf
}

func perShift(k []SupervisorListPasien, t time.Time) PerShift {
	sore := t.Add(time.Hour * 6)
	malam := t.Add(time.Hour * 12)
	var j PerShift
	var jpagi, jsore, jmalam []SupervisorListPasien
	for _, v := range k {
		if v.TglKunjungan.Before(sore) {
			jpagi = append(jpagi, v)
		} else if v.TglKunjungan.After(malam) {
			jmalam = append(jmalam, v)
		} else {
			jsore = append(jsore, v)
		}
	}
	j.Pagi = len(jpagi)
	j.Sore = len(jsore)
	j.Malam = len(jmalam)
	j.Total = (j.Pagi + j.Sore + j.Malam)
	return j
}
