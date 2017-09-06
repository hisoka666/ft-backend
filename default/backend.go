package main

import (
	ft "backend"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"cloud.google.com/go/storage"

	"google.golang.org/appengine/datastore"

	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"

	"google.golang.org/appengine/log"
)

func init() {
	http.HandleFunc("/login", login)
	http.HandleFunc("/createkursor", createKursor)
	http.HandleFunc("/createsecret", createSecret)
	http.Handle("/getsupmonth", ft.CekToken(http.HandlerFunc(getSupBulan)))

}

func getSupBulan(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	kur := &ft.KursorIGD{}
	json.NewDecoder(r.Body).Decode(kur)
	list, monini := ft.GetListSupbyKursor(ctx, kur.Bulan)
	log.Infof(ctx, "List adalah: %v", &list)

	hari, dept, shift := ft.PerHariPerBulan(ctx, &list, monini)
	log.Infof(ctx, "perhari adalah: %v", hari)
	log.Infof(ctx, "perdept adalah: %v", dept)
	log.Infof(ctx, "pershift adalah: %v", shift)
	send := &ft.SupervisorList{
		StatusServer:    "OK",
		PerHari:         hari,
		PerDeptPerHari:  dept,
		PerShiftPerHari: shift,
	}
	json.NewEncoder(w).Encode(send)

}
func createKursor(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	var staf []ft.Staff
	q := datastore.NewQuery("Staff")
	_, err := q.GetAll(ctx, &staf)
	if err != nil {
		ft.LogError(ctx, err)
	}

	for _, v := range staf {
		ft.CreateEndKursor(ctx, v.Email)
	}
	ft.CreateKursorIGD(ctx)

}

func createSecret(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	client, err := storage.NewClient(ctx)
	if err != nil {
		ft.LogError(ctx, err)
	}
	obj := client.Bucket("igdsanglah").Object("secretkey")
	wc := obj.NewWriter(ctx)
	if _, err := wc.Write(MakeKey(ctx)); err != nil {
		ft.LogError(ctx, err)
	}
	if err = wc.Close(); err != nil {
		ft.LogError(ctx, err)
	}
}

func MakeKey(ctx context.Context) []byte {
	key := make([]byte, 64)
	// ctx := appengine.NewContext(r)
	_, err := rand.Read(key)
	if err != nil {
		log.Errorf(ctx, "Error creating random number: %v", err)
	}
	return key
}

func login(w http.ResponseWriter, r *http.Request) {
	token := "https://www.googleapis.com/oauth2/v3/tokeninfo?id_token=" + r.FormValue("token")
	ctx := appengine.NewContext(r)
	client := urlfetch.Client(ctx)

	resp, err := client.Get(token)
	if err != nil {
		ft.LogError(ctx, err)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ft.LogError(ctx, err)
	}
	resp.Body.Close()

	var dat map[string]string
	if err := json.Unmarshal(b, &dat); err != nil {
		ft.LogError(ctx, err)
		return
	}

	log.Infof(ctx, dat["email"])

	user, token, peran := ft.CekStaff(ctx, dat["email"])

	if user == "no-access" {
		fmt.Fprintln(w, "no-access")
	} else if peran == "admin" {
		web := ft.AdminPage(ctx, token)
		json.NewEncoder(w).Encode(web)
	} else if peran == "supervisor" {
		web := ft.SupervisorPage(ctx, token, user)
		json.NewEncoder(w).Encode(web)
	} else {
		web := ft.GetMainContent(ctx, user, token, dat["email"])
		// js := ft.ConvertJSON(web)
		// log.Infof(ctx, string(js))
		json.NewEncoder(w).Encode(web)

	}

}
