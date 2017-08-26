package input_edit_obat

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"cloud.google.com/go/storage"

	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/search"

	"github.com/dgrijalva/jwt-go"
)

func init() {
	http.Handle("/", CekToken(http.HandlerFunc(inputEditObat)))
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

type ServerResponse struct {
	Error string `json:"error"`
}

type IndexObat struct {
	MerkDagang string `json:"merk"`
	Kandungan  string `json:"kandungan"`
	Link       string `json:"link"`
}

func inputEditObat(w http.ResponseWriter, r *http.Request) {
	link := r.URL.Path[1:]
	ctx := appengine.NewContext(r)
	log.Infof(ctx, "Link adalah: %v", link)
	obt := InputObat{}
	json.NewDecoder(r.Body).Decode(&obt)
	keyObt, _ := datastore.DecodeKey(link)
	// keyObt, _ := DatastoreKey(ctx, "InputObat", "", "", "")
	k, err := datastore.Put(ctx, keyObt, &obt)
	log.Infof(ctx, "Key k adalah: %v", k)
	if err != nil {
		m := &ServerResponse{
			Error: fmt.Sprintf("Kesalahan server: %v", err),
		}
		json.NewEncoder(w).Encode(m)
		log.Errorf(ctx, "Kesalahan server: %v", err)
	}
	oba := &IndexObat{
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
	ke, err := index.Put(ctx, k.Encode(), oba)
	if err != nil {
		log.Errorf(ctx, "Terjadi kesalahan saat memasukkan dokumen: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Infof(ctx, "Key untuk index obat adalah : %v", ke)

	m := &ServerResponse{
		Error: "",
	}
	json.NewEncoder(w).Encode(m)
}

func CekToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := appengine.NewContext(r)
		signKey, err := GetKey(ctx)
		//*email = ""

		if err != nil {
			LogError(ctx, err)
			// log.Errorf(ctx, "Error Fetching Key form Bucket: %v", err)
			return
		}

		kop := r.Header.Get("Authorization")
		log.Infof(ctx, "Token is: %v", kop)

		token, err := jwt.Parse(kop, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
			return signKey, nil
		})

		if err != nil {
			log.Errorf(ctx, "Sessions Expired: %v", err)
			//todo: fungsi untuk kembali ke halaman awal
			// http.RedirectHandler("http://localhost:9090", 303)
			LogError(ctx, err)
			fmt.Fprintln(w, "token-expired")
			return
		}
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			log.Infof(ctx, "Token Worked, issuer: %v", claims["iss"])
			//&email = claims["iss"]
		} else {
			log.Errorf(ctx, "Token not working")
			fmt.Fprintln(w, "token-not-working")
			return
		}
		//email := claims["iss"]
		next.ServeHTTP(w, r)
	})
}

func GetKey(ctx context.Context) ([]byte, error) {
	//buat context untuk akses cloud storage
	// ctx := appengine.NewContext(r)
	//buat client dari context untuk akses cloud storage
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	// harus buat object untuk akses bucket
	obj := client.Bucket("igdsanglah").Object("secretkey")
	// cek atribut Created
	// atrib, err := obj.Attrs(ctx)
	// if err != nil {
	// 	return nil, err
	// }
	// // jika Created sudah lebih dari 1 bulan, akan dibuat secretkey
	// // yang baru
	// if skrn := time.Now().After(atrib.Created.AddDate(0, 1, 0)); skrn == true {
	// 	wc := obj.NewWriter(ctx)
	// 	if _, err := wc.Write(MakeKey(ctx)); err != nil {
	// 		return nil, err
	// 	}
	// 	if err = wc.Close(); err != nil {
	// 		return nil, err
	// 	}
	// }
	// membaca secret key dari cloud storage
	rc, err := obj.NewReader(ctx)
	// jika tidak ada secretkey, akan dibuat yang baru
	// if err != nil {
	// 	wc := obj.NewWriter(ctx)
	// 	if _, err := wc.Write(MakeKey(ctx)); err != nil {
	// 		return nil, err
	// 	}
	// 	if err = wc.Close(); err != nil {
	// 		return nil, err
	// 	}
	// }
	// membaca secret key yang sudah diperoleh
	key, err := ioutil.ReadAll(rc)
	if err != nil {
		return nil, err
	}
	// jangan lupa close file
	defer rc.Close()
	// mengembalikan secret key dan error (klo ada)
	return key, nil
}

func LogError(c context.Context, e error) {
	log.Errorf(c, "Error is: %v", e)
	return
}

func DatastoreKey(ctx context.Context, kind1 string, id1 string, kind2 string, id2 string) (*datastore.Key, *datastore.Key) {
	gpKey := datastore.NewKey(ctx, "IGD", "fasttrack", 0, nil)
	parKey := datastore.NewKey(ctx, kind1, id1, 0, gpKey)
	chldKey := datastore.NewKey(ctx, kind2, id2, 0, parKey)

	return parKey, chldKey
}
