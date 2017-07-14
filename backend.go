package main

import (
	ft "backend"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"google.golang.org/appengine/datastore"

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

type InputPts struct {
	ft.DataPasien      `json:"datapts"`
	ft.KunjunganPasien `json:"kunjungan"`
}

func init() {
	http.HandleFunc("/login", login)
	http.Handle("/getcm", ft.CekToken(http.HandlerFunc(getCM)))
	http.Handle("/inputpts", ft.CekToken(http.HandlerFunc(inputPasien)))
	http.Handle("/entri/edit", ft.CekToken(http.HandlerFunc(editEntri)))
	http.Handle("/entri/confirmedit", ft.CekToken(http.HandlerFunc(editEntriConfirmed)))
	// http.HandleFunc("/getmain", mainPage)
}

func logError(c context.Context, e error) {
	// c := appengine.NewContext(r)
	log.Errorf(c, "Error is: %v", e)
	return
}

func editEntri(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	pts := ft.Pasien{}

	json.NewDecoder(r.Body).Decode(&pts)

	resp := ft.GetKunPasien(ctx, pts.LinkID)
	js := ft.ConvertJSON(resp)
	log.Infof(ctx, string(js))
	json.NewEncoder(w).Encode(resp)
}

func inputPasien(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	input := &InputPts{}
	pts := &ft.Pasien{}
	err := json.NewDecoder(r.Body).Decode(input)
	if err != nil {
		pts.NoCM = "kesalahan-decoding-json"
		json.NewEncoder(w).Encode(input)
	}

	data := &ft.DataPasien{
		NamaPasien: input.DataPasien.NamaPasien,
		TglDaftar:  input.DataPasien.TglDaftar,
	}

	kun := &ft.KunjunganPasien{
		Diagnosis:     input.KunjunganPasien.Diagnosis,
		GolIKI:        input.KunjunganPasien.GolIKI,
		ATS:           input.KunjunganPasien.ATS,
		ShiftJaga:     input.KunjunganPasien.ShiftJaga,
		JamDatang:     input.KunjunganPasien.JamDatang,
		JamDatangRiil: input.KunjunganPasien.JamDatangRiil,
		Dokter:        input.KunjunganPasien.Dokter,
		Bagian:        input.KunjunganPasien.Bagian,
	}

	numKey, inputKey := DatastoreKey(ctx, "DataPasien", input.DataPasien.NomorCM, "KunjunganPasien", "")

	if input.DataPasien.TglDaftar.IsZero() == false {
		_, err := datastore.Put(ctx, numKey, data)
		if err != nil {
			ft.LogError(ctx, err)
			pts.NoCM = "kesalahan-database"
			json.NewEncoder(w).Encode(input)
			return
		}

	}

	k, err := datastore.Put(ctx, inputKey, kun)
	if err != nil {
		ft.LogError(ctx, err)
		pts.NoCM = "kesalahan-database"
		json.NewEncoder(w).Encode(pts)
		return
	}

	pts = ft.ConvertDatastore(ctx, input.KunjunganPasien, k)

	json.NewEncoder(w).Encode(pts)

}

func UpdateEntri(w http.ResponseWriter, r *http.Request) {
	
	ubah := &UbahPasien{}
	
	err := json.NewDecoder(r.Body).Decode(ubah)
	if err != nil {
		ubah.NoCM = "kesalahan-decoding-json"
		ft.LogError(ctx, err)
		json.NewEncoder(w).Encode(ubah)
	}

	ctx := appengine.NewContext(r)
	kun := &KunjunganPasien{}
	pts := &DataPasien{}

	keyKun, err := datastore.DecodeKey(ubah.LinkID)
	if err != nil {
		ubah.NoCM = "kesalahan-decoding-Key"
		ft.LogError(ctx, err)
		json.NewEncoder(w).Encode(ubah)
	}
	keyPts := keyKun.Parent()

	err = datastore.Get(ctx, keyKun, kun)
	if err != nil {
		ubah.NoCM = "kesalahan-database-get-kunjungan-pts"
		ft.LogError(ctx, err)
		json.NewEncoder(w).Encode(ubah)
	}
	kun.Diagnosis = ubah.Diagnosis
	kun.ATS = ubah.ATS
	kun.GolIKI = ubah.IKI
	kun.ShiftJaga = ubah.Shift

	err = datastore.Get(ctx, keyPts, pts)
	if err != nil {
		ubah.NoCM = "kesalahan-database-get-datapts"
		ft.LogError(ctx, err)
		json.NewEncoder(w).Encode(ubah)return
	}
	pts.NamaPasien = ubah.NamaPasien

	if _, err := datastore.Put(ctx, keyKun, kun); err != nil {
		ubah.NoCM = "kesalahan-database-put-kunjungan-failed"
		ft.LogError(ctx, err)
		json.NewEncoder(w).Encode(ubah)
		return
	}

	if _, err := datastore.Put(ctx, keyPts, pts); err != nil {
		ubah.NoCM = "kesalahan-database-put-datapts-failed"
		ft.LogError(ctx, err)
		json.NewEncoder(w).Encode(ubah)
		return
	}

	ubah.NoCM = "OK"
	json.NewEncoder(w).Encode(ubah)
}

func DatastoreKey(ctx context.Context, kind1 string, id1 string, kind2 string, id2 string) (*datastore.Key, *datastore.Key) {
	gpKey := datastore.NewKey(ctx, "IGD", "fasttrack", 0, nil)
	parKey := datastore.NewKey(ctx, kind1, id1, 0, gpKey)
	chldKey := datastore.NewKey(ctx, kind2, id2, 0, parKey)

	return parKey, chldKey
}
func getCM(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	pts := &ft.Pasien{}

	if r.Body == nil {
		pts.NamaPasien = "request-body-empty"
		json.NewEncoder(w).Encode(pts)
	}

	err := json.NewDecoder(r.Body).Decode(&pts)
	if err != nil {
		pts.NamaPasien = "kesalahan-decoding-json"
		json.NewEncoder(w).Encode(pts)
	}

	dat, err := ft.GetNamaByNoCM(ctx, pts.NoCM)
	if err != nil {
		dat.NomorCM = "kesalahan-server"
		json.NewEncoder(w).Encode(pts)
	}
	pts.NamaPasien = dat.NamaPasien
	pts.NoCM = dat.NomorCM
	json.NewEncoder(w).Encode(pts)

}
func login(w http.ResponseWriter, r *http.Request) {
	token := "https://www.googleapis.com/oauth2/v3/tokeninfo?id_token=" + r.FormValue("token")
	ctx := appengine.NewContext(r)
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
