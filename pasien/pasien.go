package pasien

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"google.golang.org/appengine/datastore"

	"cloud.google.com/go/storage"
	jwt "github.com/dgrijalva/jwt-go"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

func init() {
	http.Handle("/", CekToken(http.HandlerFunc(getPasienDetail)))
}

type DetailPasien struct {
	Pasien    DataPasien        `json:"datapts"`
	Kunjungan []KunjunganPasien `json:"kunjungan"`
	LinkID    string            `json:"link"`
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
	NamaPasien string    `json:"namapts"`
	NomorCM    string    `json:"nocm"`
	JenKel     string    `json:"jenkel"`
	Alamat     string    `json:"alamat"`
	TglDaftar  time.Time `json:"tgldaf"`
	TglLahir   time.Time `json:"tgllhr"`
	Umur       time.Time `json:"umur"`
}

type Pasien struct {
	StatusServer string    `json:"stat"`
	TglKunjungan string    `json:"tgl"`
	ShiftJaga    string    `json:"shift"`
	ATS          string    `json:"ats"`
	Dept         string    `json:"dept"`
	NoCM         string    `json:"nocm"`
	NamaPasien   string    `json:"nama"`
	Diagnosis    string    `json:"diag"`
	IKI          string    `json:"iki"`
	LinkID       string    `json:"link"`
	TglAsli      time.Time `json:"tglasli"`
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

func getPasienDetail(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	pts := &Pasien{}
	json.NewDecoder(r.Body).Decode(pts)
	parent, _ := datastore.DecodeKey(pts.LinkID)
	parKey := parent.Parent()
	// log.Infof(ctx, "Key adalah: %v", parKey)
	q := datastore.NewQuery("KunjunganPasien").Ancestor(parKey).Filter("Hide=", false).Order("-JamDatang")
	t := q.Run(ctx)
	var n []KunjunganPasien
	for {
		var j KunjunganPasien

		k, err := t.Next(&j)
		// log.Infof(ctx, "Tgl Kunjungan adalah: %v", j.JamDatang)
		if err == datastore.Done {
			// log.Infof(ctx, "Database habis")
			break
		}

		if err != nil {
			LogError(ctx, err)
			break
		}
		j.LinkID = k.Encode()
		n = append(n, j)
	}
	var p DataPasien
	err := datastore.Get(ctx, parKey, &p)
	if err != nil {
		LogError(ctx, err)
		return
	}
	p.NomorCM = parKey.StringID()
	// log.Infof(ctx, "Nama pasien: %v", p.NamaPasien)
	dat := &DetailPasien{
		Pasien:    p,
		Kunjungan: n,
		LinkID:    parKey.Encode(),
	}

	json.NewEncoder(w).Encode(dat)
}
