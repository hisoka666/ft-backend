package backend

import (
	_ "encoding/json"
	"fmt"
	"log"
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"
)

func init() {
	http.HandleFunc("/test", test)
}

func test(w http.ResponseWriter, r *http.Request) {

	if origin := r.Header.Get("Origin"); origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		ctx := appengine.NewContext(r)
		client := urlfetch.Client(ctx)
		token := "https://www.googleapis.com/oauth2/v3/tokeninfo?id_token=" + r.FormValue("idtoken")
		resp, err := client.Get(token)
		if err != nil {
			log.Fatalf("Error Getting Token Info: %v", err)
			return
		}

		fmt.Fprintln(w, resp.Status)
	}
}
