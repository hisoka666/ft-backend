package adding_index

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/search"
)

func init() {
	// http.HandleFunc("/", index)
	// http.HandleFunc("/obat", indexObat)
	http.HandleFunc("/deleteobat/", deleteObat)
	http.HandleFunc("/igdbulan", igdBulan)
}

type KunjunganPasien struct {
	Diagnosis, LinkID      string
	GolIKI, ATS, ShiftJaga string
	JamDatang              time.Time
	Dokter                 string
	Hide                   bool
	JamDatangRiil          time.Time
	Bagian                 string
}
type DataPasien struct {
	NamaPasien, NomorCM, JenKel, Alamat string
	TglDaftar, Umur                     time.Time
}
type KursorIGD struct {
	Bulan string `json:"bulan"`
	Point string `json:"point"`
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

func igdBulan(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	q := datastore.NewQuery("KunjunganPasien").Filter("Hide=", false).Order("-JamDatang")
	t := q.Run(c)
	kur := KursorIGD{}
	kun := KunjunganPasien{}
	// days := time.Date(yr,time.Month(mo),0,0,0,0,0,zone).Day()
	// mon := time.Date()
	// zone, _ := time.LoadLocation("Asia/Makassar")
	// todayIs := time.Now().In(zone)
	// hariini := time.Date(todayIs.Year(), todayIs.Month(), 1, 8, 0, 0, 0, zone)
	// tgl := hariini.AddDate(0, -1, 0).Format("2006/01")
	// _, kurKey := DatastoreKey(c, "KursorIGD", tgl, "", "")
	// log.Infof(c, "Waktu lokal adalah: %v", hariini)
	for i := 8; i > 1; i-- {
		zone, _ := time.LoadLocation("Asia/Makassar")
		todayIs := time.Now().In(zone)
		hariini := time.Date(todayIs.Year(), time.Month(i), 1, 8, 0, 0, 0, zone)
		tgl := hariini.Format("2006/01")
		kurKey, _ := DatastoreKey(c, "KursorIGD", tgl, "", "")
		// log.Infof(c, "Key adalah: %v", kurKey)
		for {
			_, err := t.Next(&kun)
			if err == datastore.Done {
				log.Infof(c, "Data tidak ditemukan!")
				break
			}
			if err != nil {
				log.Infof(c, "Kesalahan membaca database: %v", err)
				break
				// LogError(c, err)
			}

			// jamEdit := AdjustTime(kun.JamDatang, kun.ShiftJaga)
			// log.Infof(c, "Jamedit adalah: %v", jamEdit)
			// log.Infof(c, "Apakah hari ini sebelum tanggal 1? %v", jamEdit.Before(hariini))
			if kun.JamDatang.Before(hariini) == true {
				cursor, _ := t.Cursor()
				kur.Point = cursor.String()
				kur.Bulan = tgl
				k, err := datastore.Put(c, kurKey, &kur)
				if err != nil {
					log.Errorf(c, "Kesalahan menulis database: %v", err)
					break
				}
				// log.Infof(c, "key kursor adalah: %v", k)
				// err = datastore.Get(c, k, &kur)
				// if err != nil {
				// 	log.Errorf(c, "Gagal memperoleh data : %v", err)
				// 	break
				// }
				// log.Infof(c, "Kursor adalah: %v", kur.Point)
				// log.Infof(c, "key kursor adalah: %v", k)
				log.Infof(c, "Berhasil menambahkan kursor %v", k)
				break
			}
		}
	}

}

func DatastoreKey(ctx context.Context, kind1 string, id1 string, kind2 string, id2 string) (*datastore.Key, *datastore.Key) {
	gpKey := datastore.NewKey(ctx, "IGD", "fasttrack", 0, nil)
	parKey := datastore.NewKey(ctx, kind1, id1, 0, gpKey)
	chldKey := datastore.NewKey(ctx, kind2, id2, 0, parKey)

	return parKey, chldKey
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
