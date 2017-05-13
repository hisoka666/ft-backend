package backend

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"google.golang.org/appengine"
	//urlfetch digunakan untuk mengganti http.Get dan http.Post karena tidak didukung oleh app engine
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

type Staff struct {
	Email       string
	NamaLengkap string
	LinkID      string
}

func init() {
	http.HandleFunc("/test", test)
}

func test(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	client := urlfetch.Client(ctx)
	token := "https://www.googleapis.com/oauth2/v3/tokeninfo?id_token=" + r.FormValue("idtoken")
	//resp berisi JWT dan harus diparse untuk tahu apa isinya
	resp, err := client.Get(token)
	if err != nil {
		log.Errorf(ctx, "Error Getting Token Info: %v", err)
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf(ctx, "error reading body: %v", err)
	}

	var dat map[string]string
	if err := json.Unmarshal(b, &dat); err != nil {
		panic(err)
	}

	if origin := r.Header.Get("Origin"); origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		q := datastore.NewQuery("Staff").Filter("Email =", dat["email"])
		var staf []Staff
		_, err := q.GetAll(ctx, &staf)
		if err != nil {
			fmt.Fprintln(w, "Error Fetching Data: ", err)
		}
		if len(staf) == 0 {
			fmt.Fprintln(w, "Maaf Anda tidak terdaftar sebagai staf. Mohon hubungi Admin")
		}

		for _, v := range staf {
			fmt.Fprintln(w, "Selamat Datang, ", v.NamaLengkap)
			fmt.Fprintln(w, string(Last100(r, v.Email)))
		}
	}
}

type EntriPasien struct {
	Tgl, Jaga, NoCM, NamaPts, Diagnosis, IKI, Entri string
}
type DataPasien struct {
	NamaPasien, NomorCM, JenKel, Alamat string
	TglDaftar, Umur                     time.Time
}
type KunjunganPasien struct {
	Diagnosis, LinkID, GolIKI, ATS, ShiftJaga, Dokter string
	JamDatang, JamDatangRiil                          time.Time
	Hide                                              bool
}

func Last100(r *http.Request, email string) []byte {
	ctx := appengine.NewContext(r)
	q := datastore.NewQuery("KunjunganPasien").Limit(100).Filter("Dokter =", email).Order("-JamDatang")
	var m []EntriPasien
	t := q.Run(ctx)
	for {
		var j KunjunganPasien
		k, err := t.Next(&j)
		if err == datastore.Done {
			break
		}
		if err != nil {
			log.Errorf(ctx, "fetching next Person: %v", err)
			break
		}
		tanggal := UbahTanggal(j.JamDatang, j.ShiftJaga)
		nocm, namapts := GetDataPts(r, k)
		//di sini menggunakan pointer untuk *n
		n := &EntriPasien{
			Tgl:       tanggal,
			Jaga:      j.ShiftJaga,
			NoCM:      nocm,
			NamaPts:   namapts,
			Diagnosis: j.Diagnosis,
			IKI:       j.GolIKI,
			Entri:     k.Encode(),
		}
		m = append(m, *n)
	}

	jsm, err := json.Marshal(m)
	if err != nil {
		log.Errorf(ctx, "error marshalling json: %v", err)
	}
	return jsm
}

func UbahTanggal(tgl time.Time, shift string) string {
	jam := tgl.Hour()
	tglstring := ""
	if jam < 12 && shift == "3" {
		tglstring = tgl.AddDate(0, 0, -1).Format("02-01-2006")
	} else {
		tglstring = tgl.Format("02-01-2006")
	}
	return tglstring
}

func GetDataPts(r *http.Request, k *datastore.Key) (no, nama string) {
	ctx := appengine.NewContext(r)
	var p DataPasien
	keypar := k.Parent()
	err := datastore.Get(ctx, keypar, &p)
	if err != nil {
		log.Errorf(ctx, "error getting data: %v", err)
	}
	no = keypar.StringID()
	nama = p.NamaPasien
	return no, nama
}

// func GetListPasien(r *http.Request, email string, m, y int) []ListPasien {
// 	ctx := appengine.NewContext(r)
// 	monIn := DatebyInt(m, y)
// 	q := datastore.NewQuery("KunjunganPasien").Filter("Dokter =", email).Filter("Hide =", false).Order("-JamDatang")
// 	list := IterateList(ctx, w, q, monIn)
// 	return list
// }
//
// func IterateList(ctx appengine.Context, w http.ResponseWriter, q *datastore.Query, mon time.Time) []ListPasien {
// 	t := q.Run(ctx)
// 	monAf := mon.AddDate(0, 1, 0)
// 	var daf KunjunganPasien
// 	var tar ListPasien
// 	var pts DataPasien
// 	var list []ListPasien
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
//
// 		nocm := k.Parent()
// 		tar.NomorCM = nocm.StringID()
//
// 		err = datastore.Get(ctx, nocm, &pts)
// 		if err != nil {
// 			fmt.Fprintln(w, "Error Fetching Data Pasien: ", err)
// 		}
//
// 		tar.NamaPasien = ProperTitle(pts.NamaPasien)
// 		tar.Diagnosis = ProperTitle(daf.Diagnosis)
// 		tar.ShiftJaga = daf.ShiftJaga
// 		tar.LinkID = k.Encode()
//
// 		if daf.GolIKI == "1" {
// 			tar.IKI1 = "1"
// 			tar.IKI2 = ""
// 		} else {
// 			tar.IKI1 = ""
// 			tar.IKI2 = "1"
// 		}
//
// 		list = append(list, tar)
// 	}
//
// 	return list
// }
