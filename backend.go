package main

import (
	ft "backend"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"

	"google.golang.org/appengine/log"
)

func init() {
	http.HandleFunc("/login", login)
}

func logError(c context.Context, e error) {
	// c := appengine.NewContext(r)
	log.Errorf(c, "Error is: %v", e)
	return
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
		log.Infof(ctx, string(ft.GetMainContent(ctx, user, token, dat["email"])))
		fmt.Fprintln(w, token)
	}
	// jwt := ft.ConvertJSON(ft.CekStaff(ctx, dat["email"]))
	// log.Infof(ctx, string(jwt))
	// w.Header().Set("Content-Type", "application/json")
	// w.Write(jwt)
}
