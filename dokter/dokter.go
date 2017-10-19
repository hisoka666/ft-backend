package dokter

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"cloud.google.com/go/storage"
	jwt "github.com/dgrijalva/jwt-go"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

func init() {
	http.Handle("/", CekToken(http.HandlerFunc(getDocData)))
	http.Handle("/simpan-dokter", CekToken(http.HandlerFunc(simpanDocData)))
}

type DetailStaf struct {
	NamaLengkap  string    `json:"nama"`
	NIP          string    `json:"nip"`
	NPP          string    `json:"npp"`
	GolonganPNS  string    `json:"golpns"`
	Alamat       string    `json:"alamat"`
	Bagian       string    `json:"bagian"`
	LinkID       string    `json:"link"`
	TanggalLahir time.Time `json:"tgl"`
	Umur         string    `json:"umur"`
}
type Staff struct {
	Email, NamaLengkap, LinkID, Peran string
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
	// membaca secret key dari cloud storage
	rc, err := obj.NewReader(ctx)
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

func getDocData(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	doc := &Staff{}
	json.NewDecoder(r.Body).Decode(doc)
	k, err := datastore.DecodeKey(doc.LinkID)
	if err != nil {
		LogError(ctx, err)
	}
	q := datastore.NewQuery("DetailStaf").Ancestor(k)
	t := q.Run(ctx)
	det := &DetailStaf{}
	for {
		_, err = t.Next(det)
		if err == datastore.Done {
			err = datastore.Get(ctx, k, doc)
			if err != nil {
				LogError(ctx, err)
			}
			detail := &DetailStaf{
				NamaLengkap: doc.NamaLengkap,
				LinkID:      doc.LinkID,
			}
			log.Infof(ctx, "Data belum ada, sementara pake ini %v", detail)
			json.NewEncoder(w).Encode(detail)
			break
		}
		if err != nil {
			LogError(ctx, err)
			break
		}
		log.Infof(ctx, "data ada, yang kirim adalah : %v", det)
		json.NewEncoder(w).Encode(det)
		break
	}
}

func simpanDocData(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	det := &DetailStaf{}
	json.NewDecoder(r.Body).Decode(det)
	parent := &det.LinkID
	keyPar, err := datastore.DecodeKey(*parent)
	if err != nil {
		LogError(ctx, err)
	}
	q := datastore.NewQuery("DetailStaf").Ancestor(keyPar)
	t := q.Run(ctx)
	detail := &DetailStaf{}
	for {
		k, err := t.Next(detail)
		if err == datastore.Done {
			keyDoc := datastore.NewKey(ctx, "DetailStaf", *parent, 0, keyPar)
			det.LinkID = keyDoc.Encode()
			_, err = datastore.Put(ctx, keyDoc, det)
			if err != nil {
				LogError(ctx, err)
			}
			log.Infof(ctx, "Berhasil memasukkan data dokter baru")
			json.NewEncoder(w).Encode(det)
			break
		}
		if err != nil {
			LogError(ctx, err)
			break
		}
		det.LinkID = k.Encode()
		_, err = datastore.Put(ctx, k, det)
		if err != nil {
			LogError(ctx, err)
		}
		log.Infof(ctx, "berhasil mengubah data lama")
		json.NewEncoder(w).Encode(det)
	}
	log.Infof(ctx, "isi data adalah: %v", det)
	// log.Infof(ctx, "keyparent adalah: %v", *parent)
	// keyDoc := datastore.NewKey(ctx, "DetailStaf", *parent, 0, keyPar)
	// det.LinkID = keyDoc.Encode()
	// log.Infof(ctx, "Isi data dokter : %v", det)
	// _, err = datastore.Put(ctx, keyDoc, det)
	// if err != nil {
	// 	LogError(ctx, err)
	// }
	// det.LinkID = *parent
	// json.NewEncoder(w).Encode(det)
}
