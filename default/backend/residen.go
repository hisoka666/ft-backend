package backend

import (
	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

func ResidenPage(c context.Context, user, token, peran, link string) *MainView {
	dept := peran[8:]
	g := &MainView{
		User:   user,
		Token:  token,
		LinkID: link,
		Peran:  peran,
	}
	log.Infof(c, "Residen dari bagian: %v", dept)
	pts := []Pasien{}
	q := datastore.NewQuery("KunjunganPasien").Order("-JamDatang").Filter("Hide =", false)
	t := q.Run(c)
	for {
		var kun KunjunganPasien
		k, err := t.Next(&kun)
		log.Infof(c, "Bagian adalah: %v", kun.Bagian)
		if err != nil {
			break
		}
		if kun.Bagian != dept {
			continue
		}
		if len(pts) >= 50 {
			break
		}
		pas := ConvertDatastore(c, kun, k)
		log.Infof(c, "Nama Pasien adalah: %v", pas.NamaPasien)
		pts = append(pts, *pas)
	}
	g.Pasien = pts
	log.Infof(c, "List pasien ini adalah : %v", g.Pasien)
	return g
}
