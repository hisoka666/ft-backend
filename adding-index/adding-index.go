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
	http.HandleFunc("/obat", indexObat)
	http.HandleFunc("/deleteobat/", deleteObat)
}

type DataPasien struct {
	NamaPasien, NomorCM, JenKel, Alamat string
	TglDaftar, Umur                     time.Time
}

type IndexDataPasien struct {
	Nama string
	NoCM string
}

type IndexObat struct {
	MerkDagang string `json:"merk"`
	Kandungan  string `json:"kandungan"`
	Link       string `json:"link"`
}

type InputObat struct {
	MerkDagang     string   `json:"merk"`
	Kandungan      string   `json:"kand"`
	MinDose        string   `json:"mindose"`
	MaxDose        string   `json:"maxdose"`
	Tablet         []string `json:"tab"`
	Sirup          []string `json:"syr"`
	Drop           []string `json:"drop"`
	Lainnya        string   `json:"lainnya"`
	SediaanLainnya []string `json:"lainnya_sediaan"`
	Rekomendasi    string   `json:"rekom"`
	Dokter         string   `json:"doc"`
}

func deleteObat(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	index, err := search.Open("DataObat")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id := r.URL.Path[12:]
	err = index.Delete(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, "Deleted document: ", id)
}
func indexObat(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	q := datastore.NewQuery("InputObat")
	t := q.Run(ctx)
	for {
		var obt InputObat
		k, err := t.Next(&obt)
		if err == datastore.Done {
			log.Infof(ctx, "Tidak ada data lagi")
			break
		}
		if err != nil {
			log.Infof(ctx, "Mengambil data selanjutnya: %v", err)
			break
		}

		ind := &IndexObat{
			MerkDagang: obt.MerkDagang,
			Kandungan:  obt.Kandungan,
			Link:       k.Encode(),
		}

		index, err := search.Open("DataObat")
		if err != nil {
			log.Errorf(ctx, "Terjadi kesalahan saat membuat dokumen: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		ke, err := index.Put(ctx, k.Encode(), ind)
		if err != nil {
			log.Errorf(ctx, "Terjadi kesalahan saat memasukkan dokumen: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Infof(ctx, "Key untuk index obat adalah : %v", ke)
	}
	fmt.Fprintln(w, "all is well")
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
