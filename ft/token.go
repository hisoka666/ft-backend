package ft

import (
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"cloud.google.com/go/storage"
	"github.com/dgrijalva/jwt-go"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

// HTTP Middleware untuk mengecek token tiap ada request

func CekToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := appengine.NewContext(r)
		signKey, err := GetKey(r)
		//*email = ""

		if err != nil {
			log.Errorf(ctx, "Error Fetching Key form Bucket: %v", err)
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
			http.RedirectHandler("http://localhost:4200/login", 303)
			return
		}
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			log.Infof(ctx, "Token Worked, issuer: %v", claims["iss"])
			//&email = claims["iss"]
		} else {
			log.Errorf(ctx, "Token not working")
		}
		//email := claims["iss"]
		next.ServeHTTP(w, r)
	})
}

//fungsi untuk membuat token, token diambil dari cloud store, kemudian
//digunakan dalam metode jwt untuk menghasilkan token. nantinya secret
//di cloud store akan diupdate setiap bulan
func CreateToken(w http.ResponseWriter, r *http.Request, email string) string {
	//mengambil secretkey dari cloud storage
	secret, err := GetKey(r)
	if err != nil {
		fmt.Fprintf(w, "Error Fetching Bucket: %v", err)
	}
	claims := &jwt.StandardClaims{
		//mengeset expiration date untuk token
		ExpiresAt: time.Now().Add(time.Hour * 12).Unix(),
		Issuer:    email,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tok, err := token.SignedString(secret)
	if err != nil {
		fmt.Fprintln(w, err)
	}
	return tok
}

//fungsi ini mengambil secret key di cloud storage
//sekaligus mengecek apakah kuncinya update
//
func GetKey(r *http.Request) ([]byte, error) {
	//buat context untuk akses cloud storage
	ctx := appengine.NewContext(r)
	//buat client dari context untuk akses cloud storage
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	// harus buat object untuk akses bucket
	obj := client.Bucket("igdsanglah").Object("secretkey")
	// cek atribut Created
	atrib, err := obj.Attrs(ctx)
	if err != nil {
		return nil, err
	}
	// jika Created sudah lebih dari 1 bulan, akan dibuat secretkey
	// yang baru
	if skrn := time.Now().After(atrib.Created.AddDate(0, 1, 0)); skrn == true {
		wc := obj.NewWriter(ctx)
		if _, err := wc.Write(MakeKey(r)); err != nil {
			return nil, err
		}
		if err = wc.Close(); err != nil {
			return nil, err
		}
	}
	// membaca secret key dari cloud storage
	rc, err := obj.NewReader(ctx)
	// jika tidak ada secretkey, akan dibuat yang baru
	if err != nil {
		wc := obj.NewWriter(ctx)
		if _, err := wc.Write(MakeKey(r)); err != nil {
			return nil, err
		}
		if err = wc.Close(); err != nil {
			return nil, err
		}
	}
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

// fungsi untuk membuat secret key yang nantinya akan disimpan
// di cloud storage
func MakeKey(r *http.Request) []byte {
	key := make([]byte, 64)
	ctx := appengine.NewContext(r)
	_, err := rand.Read(key)
	if err != nil {
		log.Errorf(ctx, "Error creating random number: %v", err)
	}
	return key
}
