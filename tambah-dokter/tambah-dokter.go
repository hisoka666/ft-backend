package tambah_dokter

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"cloud.google.com/go/storage"
	jwt "github.com/dgrijalva/jwt-go"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

func init() {
	http.Handle("/", CekToken(http.HandlerFunc(tambahDokter)))
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

type Staff struct {
	Email, NamaLengkap, LinkID, Peran string
}

func tambahDokter(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	st := &Staff{}
	json.NewDecoder(r.Body).Decode(st)

	parent, _ := DatastoreKey(ctx, "Staff", "", "", "")

	k, err := datastore.Put(ctx, parent, st)
	if err != nil {
		LogError(ctx, err)
		st.Email = "Kesalahan dalam menginput dokter"
		json.NewEncoder(w).Encode(st)
		return
	}

	st.LinkID = k.Encode()
	json.NewEncoder(w).Encode(st)
}

func DatastoreKey(ctx context.Context, kind1 string, id1 string, kind2 string, id2 string) (*datastore.Key, *datastore.Key) {
	gpKey := datastore.NewKey(ctx, "IGD", "fasttrack", 0, nil)
	parKey := datastore.NewKey(ctx, kind1, id1, 0, gpKey)
	chldKey := datastore.NewKey(ctx, kind2, id2, 0, parKey)

	return parKey, chldKey
}
