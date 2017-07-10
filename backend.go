package main

import (
	ft "backend"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"appengine/datastore"

	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"

	"google.golang.org/appengine/log"
)

type Pasien struct {
	TglKunjungan string `json:"tgl"`
	ShiftJaga    string `json:"shift"`
	NoCM         string `json:"nocm"`
	NamaPasien   string `json:"nama"`
	Diagnosis    string `json:"diag"`
	IKI1         string `json:"iki1"`
	IKI2         string `json:"iki2"`
	LinkID       string `json:"link"`
}

func init() {
	http.HandleFunc("/login", login)
	http.Handle("/getcm", ft.CekToken(http.HandlerFunc(getCM)))
	http.Handle("/inputpts", ft.CekToken(http.HandlerFunc(inputPasien)))
	// http.HandleFunc("/getmain", mainPage)
}

func logError(c context.Context, e error) {
	// c := appengine.NewContext(r)
	log.Errorf(c, "Error is: %v", e)
	return
}

// func homePage(w http.ResponseWriter, r *http.Request, email string) {
// 	// token := "https://www.googleapis.com/oauth2/v3/tokeninfo?id_token=" + r.FormValue("token")
// 	// ctx := appengine.NewContext(r)
// 	// // log.Infof(ctx, "Token adalah: %v", r.FormValue("token"))
// 	// client := urlfetch.Client(ctx)

// 	// resp, err := client.Get(token)
// 	// if err != nil {
// 	// 	logError(ctx, err)
// 	// }

// 	// b, err := ioutil.ReadAll(resp.Body)
// 	// if err != nil {
// 	// 	logError(ctx, err)
// 	// }
// 	// resp.Body.Close()

// 	// var dat map[string]string
// 	// if err := json.Unmarshal(b, &dat); err != nil {
// 	// 	logError(ctx, err)
// 	// 	return
// 	// }

// 	// log.Infof(ctx, dat["email"])

// 	// user, token := ft.CekStaff(ctx, dat["email"])

// 	// if user == "no-access" {
// 	// 	fmt.Fprintln(w, "no-access")
// 	// } else {
// 	web := ft.GetMainContent(ctx, user, token, email)
// 	js := ft.ConvertJSON(web)
// 	log.Infof(ctx, string(js))
// 	json.NewEncoder(w).Encode(web)

// 	// }
// }

func inputPts(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	input := &InputPts{}

	err := json.NewDecoder(r.Body).Decode(input)
	if err != nil {
		input.DataPasien.NamaPasien = "kesalahan-decoding-json"
		json.NewEncoder(w).Encode(input)
	}

	// 	ctx := appengine.NewContext(r)
	// nocm := r.FormValue("nocm")
	// doc, _, _ := AppCtx(ctx, "", "", "", "")
	// _, parentKey, pasienKey := AppCtx(ctx, "DataPasien", nocm, "KunjunganPasien", "")
	nocm := input.DataPasien.NoCM
	data := &DataPasien{
		NamaPasien: input.DataPasien.NamaPasien,
		TglDaftar:  input.DataPasien.TglDaftar,
	}

	kun := &KunjunganPasien{
		Diagnosis:     input.KunjunganPasien.Diagnosis,
		GolIKI:        input.KunjunganPasien.GolIKI,
		ATS:           input.KunjunganPasien.ATS,
		ShiftJaga:     input.KunjunganPasien.ShiftJaga,
		JamDatang:     input.KunjunganPasien.JamDatang,
		JamDatangRiil: input.KunjunganPasien.JamDatangRiil,
		Dokter:        input.KunjunganPasien.Dokter,
	}
	//Todo: bikin key
	// parKey := datastore.NewKey(c, "IGD", "fasttrack", 0, nil)
	// ptsKey := datastore.NewKey(c, "DataPasien", nocm, 0, parKey)

	if input.DataPasien.TglDaftar != nil {
		if _, err := datastore.Put(ctx, parentKey, data); err != nil {
			ft.logError(ctx, "Error Database: %v", err)
			return
		}
		if _, err := datastore.Put(ctx, pasienKey, kun); err != nil {
			ft.logError(ctx, "Error Database: %v", err)
			return
		}
	} else {
		if _, err := datastore.Put(ctx, pasienKey, kun); err != nil {
			ft.logError(ctx, "Error Database: %v", err)
			return
		}

	}

	time.Sleep(5000 * time.Millisecond)
	http.Redirect(w, r, "/mainpage", http.StatusSeeOther)
}
func getCM(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	var pts Pasien

	if r.Body == nil {
		pts.NamaPasien = "request-body-empty"
		json.NewEncoder(w).Encode(pts)
	}

	err := json.NewDecoder(r.Body).Decode(&pts)
	if err != nil {
		pts.NamaPasien = "kesalahan-decoding-json"
		json.NewEncoder(w).Encode(pts)
	}

	pts = ft.GetNamaByNoCM(ctx, pts.NoCM)
	if pts.NamaPasien != nil {
		pts.Baru = false
	} else {
		pts.Baru = true
	}
	json.NewEncoder(w).Encode(pts)

}
func login(w http.ResponseWriter, r *http.Request) {
	token := "https://www.googleapis.com/oauth2/v3/tokeninfo?id_token=" + r.FormValue("token")
	ctx := appengine.NewContext(r)
	// log.Infof(ctx, "Token adalah: %v", r.FormValue("token"))
	client := urlfetch.Client(ctx)

	resp, err := client.Get(token)
	if err != nil {
		logError(ctx, err)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logError(ctx, err)
	}
	resp.Body.Close()

	var dat map[string]string
	if err := json.Unmarshal(b, &dat); err != nil {
		logError(ctx, err)
		return
	}

	log.Infof(ctx, dat["email"])

	user, token := ft.CekStaff(ctx, dat["email"])

	if user == "no-access" {
		fmt.Fprintln(w, "no-access")
	} else {
		web := ft.GetMainContent(ctx, user, token, dat["email"])
		js := ft.ConvertJSON(web)
		log.Infof(ctx, string(js))
		json.NewEncoder(w).Encode(web)

	}

}
