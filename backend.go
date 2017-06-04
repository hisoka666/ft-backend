package backend

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"cloud.google.com/go/storage"

	"github.com/dgrijalva/jwt-go"
	"google.golang.org/appengine"
	//urlfetch digunakan untuk mengganti http.Get dan http.Post karena tidak didukung oleh app engine
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

type Staff struct {
	Email       string
	NamaLengkap string
	LinkID      string
}

func init() {
	http.Handle("/testuser", cekToken(http.HandlerFunc(hello)))
	http.HandleFunc("/test", test)
}

// HTTP Middleware untuk mengecek token tiap ada request
func cekToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := appengine.NewContext(r)
		signKey, err := getKey(r)

		if err != nil {
			log.Errorf(ctx, "Error Fetching Key form Bucket: %v", err)
			return
		}

		kop := r.Header.Get("Authorization")

		token, err := jwt.Parse(kop, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
			return signKey, nil
		})

		if err != nil {
			log.Errorf(ctx, "Sessions Expired: %v", err)
			http.RedirectHandler("https://www.google.co.id", 303)
			return
		}
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			log.Infof(ctx, "Token Worked, issuer: %v", claims["iss"])
		} else {
			log.Errorf(ctx, "Token not working")
		}

		next.ServeHTTP(w, r)
	})
}

//fungsi ini mengambil secret key di cloud storage
//sekaligus mengecek apakah kuncinya update
//
func getKey(r *http.Request) ([]byte, error) {
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
		if _, err := wc.Write(makeKey(r)); err != nil {
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
		if _, err := wc.Write(makeKey(r)); err != nil {
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
func makeKey(r *http.Request) []byte {
	key := make([]byte, 64)
	ctx := appengine.NewContext(r)
	_, err := rand.Read(key)
	if err != nil {
		log.Errorf(ctx, "Error creating random number: %v", err)
	}
	return key
}
func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello World")
}
func test(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	//http.Get versi appengine
	client := urlfetch.Client(ctx)
	token := "https://www.googleapis.com/oauth2/v3/tokeninfo?id_token=" + r.FormValue("idtoken")
	//resp (Response) berisi respon dari google setelah di-autentikasi melalui tokeninfo
	resp, err := client.Get(token)
	if err != nil {
		log.Errorf(ctx, "Error Getting Token Info: %v", err)
		return
	}
	//hanya resp.Body yang dibaca dan sudah berisi string dari hasil autentikasi
	//melalui tokeninfo (email, nama, dll)
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf(ctx, "error reading body: %v", err)
	}
	//jangan lupa menutup resp.Body
	resp.Body.Close()

	//sekarang mengubah string yang diperoleh dari resp.Body ke JSON untuk mempermudah
	//mengekstrak email yang akan digunakan untuk mengecek apakah sudah masuk staff
	//atau belum
	var dat map[string]string
	if err := json.Unmarshal(b, &dat); err != nil {
		panic(err)
	}

	//harus diset Header menjadi Access-Control-Allow-ORigin
	if origin := r.Header.Get("Origin"); origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers",
			"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		//Mencari di database alamat email
		q := datastore.NewQuery("Staff").Filter("Email =", dat["email"])
		var staf []Staff
		_, err := q.GetAll(ctx, &staf)
		if err != nil {
			fmt.Fprintln(w, "Error Fetching Data: ", err)
		}
		if len(staf) == 0 {
			fmt.Fprintln(w, "Maaf Anda tidak terdaftar sebagai staf. Mohon hubungi Admin")
		}

		// key, err := getKey(r)
		// if err != nil {
		// 	log.Errorf(ctx, "Error reading to bucket: %v", err)
		// 	return
		// }
		//Setelah ketemu, membuat respon untuk Ajax
		for _, v := range staf {
			//kirim token ke browser/aplikasi
			fmt.Fprintln(w, CreateToken(w, r, v.Email))
		}
	}
}

type EntriPasien struct {
	Tgl, Jaga, NoCM, NamaPts, Diagnosis, IKI, Entri string
}
type DataPasien struct {
	NamaPasien, NomorCM, JenKel, Alamat string
	TglDaftar, Umur                     time.Time
}
type KunjunganPasien struct {
	Diagnosis, LinkID, GolIKI, ATS, ShiftJaga, Dokter string
	JamDatang, JamDatangRiil                          time.Time
	Hide                                              bool
}

//fungsi untuk membuat token, token diambil dari cloud store, kemudian
//digunakan dalam metode jwt untuk menghasilkan token. nantinya secret
//di cloud store akan diupdate setiap bulan
func CreateToken(w http.ResponseWriter, r *http.Request, email string) string {
	//mengambil secretkey dari cloud storage
	secret, err := getKey(r)
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

func Last100(r *http.Request, email string) []byte {
	ctx := appengine.NewContext(r)
	q := datastore.NewQuery("KunjunganPasien").Limit(100).Filter("Dokter =", email).Order("-JamDatang")
	var m []EntriPasien
	t := q.Run(ctx)
	for {
		var j KunjunganPasien
		k, err := t.Next(&j)
		if err == datastore.Done {
			break
		}
		if err != nil {
			log.Errorf(ctx, "fetching next Person: %v", err)
			break
		}
		tanggal := UbahTanggal(j.JamDatang, j.ShiftJaga)
		nocm, namapts := GetDataPts(r, k)
		//di sini menggunakan pointer untuk *n
		n := &EntriPasien{
			Tgl:       tanggal,
			Jaga:      j.ShiftJaga,
			NoCM:      nocm,
			NamaPts:   namapts,
			Diagnosis: j.Diagnosis,
			IKI:       j.GolIKI,
			Entri:     k.Encode(),
		}
		m = append(m, *n)
	}

	jsm, err := json.Marshal(m)
	if err != nil {
		log.Errorf(ctx, "error marshalling json: %v", err)
	}
	return jsm
}

func UbahTanggal(tgl time.Time, shift string) string {
	jam := tgl.Hour()
	tglstring := ""
	if jam < 12 && shift == "3" {
		tglstring = tgl.AddDate(0, 0, -1).Format("02-01-2006")
	} else {
		tglstring = tgl.Format("02-01-2006")
	}
	return tglstring
}

func GetDataPts(r *http.Request, k *datastore.Key) (no, nama string) {
	ctx := appengine.NewContext(r)
	var p DataPasien
	keypar := k.Parent()
	err := datastore.Get(ctx, keypar, &p)
	if err != nil {
		log.Errorf(ctx, "error getting data: %v", err)
	}
	no = keypar.StringID()
	nama = p.NamaPasien
	return no, nama
}
