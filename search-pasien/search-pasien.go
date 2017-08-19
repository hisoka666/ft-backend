package search_pasien

import (
	"fmt"
	"html/template"
	lg "log"
	"net/http"

	"google.golang.org/appengine/log"

	"google.golang.org/appengine"
	"google.golang.org/appengine/search"
)

func init() {
	http.HandleFunc("/", index)
	http.HandleFunc("/caripts", cari)
}

func index(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.New("index.html").ParseFiles("index.html"))
	err := tmpl.Execute(w, nil)
	if err != nil {
		lg.Fatalf("Error adalah: %v", err)
	}
}

type IndexDataPasien struct {
	Nama string
	NoCM string
}

func cari(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	index, err := search.Open("DataPasienTest02")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	namapts := r.FormValue("nama")
	que := "snippet('" + r.FormValue("nama") + "', Nama, 50)"
	log.Infof(ctx, "String nama adalah : %s", que)
	// log.Infof(ctx, "Nama adalah: %v", namauser)
	// qu := "Nama: " + nama
	fieldex := search.FieldExpression{
		Name: "snippetList",
		Expr: que,
	}
	expre := &search.SearchOptions{
		// Limit:         0,
		// IDsOnly:       false,
		// Sort:          nil,
		// Fields:      []string{"Nama"},
		Expressions: []search.FieldExpression{fieldex},
		// Facets:        nil,
		// Refinements:   nil,
		// Cursor:        "",
		// Offset:        0,
		// CountAccuracy: 0,
	}
	q := "Nama = " + namapts
	t := index.Search(ctx, q, expre)
	for {
		var us IndexDataPasien
		_, err := t.Next(&us)
		if err == search.Done {
			break
		}
		if err != nil {
			log.Errorf(ctx, "Tidak bisa mencari: %v", err)
			break
		}

		fmt.Fprintf(w, "<li> %v </li>", us.Nama)
	}
}
