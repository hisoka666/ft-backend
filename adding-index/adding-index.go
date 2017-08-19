package adding_index

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"google.golang.org/appengine/search"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

func init() {
	http.HandleFunc("/", index)
}

type DataPasien struct {
	NamaPasien, NomorCM, JenKel, Alamat string
	TglDaftar, Umur                     time.Time
}

type IndexDataPasien struct {
	Nama string
	NoCM string
}

func index(w http.ResponseWriter, r *http.Request) {
	nextnum, _ := strconv.Atoi(r.URL.Path[1:])

	ctx := appengine.NewContext(r)
	// var dat DataPasien
	q := datastore.NewQuery("DataPasien").Offset(nextnum).Limit(500)
	t := q.Run(ctx)
	for {
		var dat DataPasien
		k, err := t.Next(&dat)
		if err == datastore.Done {
			log.Infof(ctx, "Tidak ada data lagi")
			break
		}
		if err != nil {
			log.Infof(ctx, "Mengambil data selanjutnya: %v", err)
			break
		}
		ind := &IndexDataPasien{
			Nama: dat.NamaPasien,
			NoCM: k.StringID(),
		}
		index, err := search.Open("DataPasienTest02")
		if err != nil {
			log.Infof(ctx, "Tidak bisa membuat index: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = index.Put(ctx, k.Encode(), ind)
		if err != nil {
			log.Infof(ctx, "Terjadi kesalahan dalam menyimpan index: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	}
	fmt.Fprintf(w, "Next offset is %v", nextnum+500)
}
