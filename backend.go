package main

import (
	ft "backend"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

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
	http.Handle("/entri/confirmedit", ft.CekToken(http.HandlerFunc(UpdateEntri)))
	http.Handle("/entri/delentri", ft.CekToken(http.HandlerFunc(delEntri)))
	http.Handle("/entri/firstitems", ft.CekToken(http.HandlerFunc(firstItems)))
	http.Handle("/entri/ubahtanggal", ft.CekToken(http.HandlerFunc(editDate)))
	http.Handle("/entri/confubahtanggal", ft.CekToken(http.HandlerFunc(confEditDate)))

	// http.HandleFunc("/getmain", mainPage)
}

func logError(c context.Context, e error) {
	// c := appengine.NewContext(r)
	log.Errorf(c, "Error is: %v", e)
	return
}

func firstItems(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	rec := &ft.MainView{}
	json.NewDecoder(r.Body).Decode(rec)
	rec.Pasien = ft.GetLast100(ctx, rec.User)
	js := ft.ConvertJSON(rec.Pasien)
	log.Infof(ctx, "Email adalah : %v dan list pasien adalah: %v", rec.User, string(js))

	json.NewEncoder(w).Encode(rec)
}

//////////////////////////////////////////////////////////////////////////

func confEditDate(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	pts := ft.Pasien{}

	json.NewDecoder(r.Body).Decode(&pts)
	log.Infof(ctx, pts.LinkID)
	keyKun, err := datastore.DecodeKey(pts.LinkID)
	if err != nil {
		ft.LogError(ctx, err)
	}
	log.Infof(ctx, "Key adalah : %v", keyKun)
	kun := ft.KunjunganPasien{}

	err = datastore.Get(ctx, keyKun, &kun)
	if err != nil {
		logError(ctx, err)
	}
	// log.Infof(ctx, pts.TglKunjungan)
	email := kun.Dokter
	zone, _ := time.LoadLocation("Asia/Makassar")
	jamlama := kun.JamDatang.In(zone)
	strJam := jamlama.Format("01/02/2006 15:04:05")
	jam := pts.TglKunjungan + strJam[10:]
	newTime, _ := time.ParseInLocation("2006-01-02 15:04:05", jam, zone)
	log.Infof(ctx, "Jam baru adalah: %v", newTime)

	kun.JamDatang = newTime

	_, err = datastore.Put(ctx, keyKun, &kun)
	if err != nil {
		ft.LogError(ctx, err)
		pts.NoCM = "kesalahan-database"
		json.NewEncoder(w).Encode(pts)
		return
	}
	list := ft.GetLast100(ctx, email)
	send := ft.MainView{
		Token:  "OK",
		Pasien: list,
	}
	json.NewEncoder(w).Encode(send)
	//todo: ubahTanggalPasien

}

//////////////////////////////////////////////////////////////////////////
func editDate(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	pts := ft.Pasien{}

	json.NewDecoder(r.Body).Decode(&pts)

	resp := ft.GetKunPasien(ctx, pts.LinkID)

	log.Infof(ctx, "Mengambil data pasien untuk diubah tanggalnya")
	json.NewEncoder(w).Encode(resp)
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

func delEntri(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	del := &ft.Pasien{}

	json.NewDecoder(r.Body).Decode(del)

	res := ft.DeleteEntri(ctx, del.LinkID)

	json.NewEncoder(w).Encode(res)

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
	ctx := appengine.NewContext(r)
	ubah := &ft.Pasien{}

	err := json.NewDecoder(r.Body).Decode(ubah)
	if err != nil {
		ubah.StatusServer = "kesalahan-decoding-json"
		ft.LogError(ctx, err)
		json.NewEncoder(w).Encode(ubah)
	}

	up, err := ft.UpdateEntri(ctx, ubah)
	if err != nil {
		ft.LogError(ctx, err)
		json.NewEncoder(w).Encode(up)
		return
	}

	json.NewEncoder(w).Encode(up)
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
		// js := ft.ConvertJSON(web)
		// log.Infof(ctx, string(js))
		json.NewEncoder(w).Encode(web)

	}

}
