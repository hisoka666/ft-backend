package ubah_tanggal

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
	http.Handle("/", CekToken(http.HandlerFunc(editDate)))
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

func editDate(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	pts := Pasien{}

	json.NewDecoder(r.Body).Decode(&pts)

	resp := GetKunPasien(ctx, pts.LinkID)

	log.Infof(ctx, "Mengambil data pasien untuk diubah tanggalnya")
	json.NewEncoder(w).Encode(resp)
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

func GetKunPasien(c context.Context, link string) *Pasien {
	var kun KunjunganPasien
	var dat DataPasien

	log.Infof(c, "Link adalah : %v", link)
	keyKun, err := datastore.DecodeKey(link)
	if err != nil {
		LogError(c, err)
	}
	err = datastore.Get(c, keyKun, &kun)
	if err != nil {
		LogError(c, err)
	}
	// log.Infof(c, "Diagnosis adalah: %v", kun.Diagnosis)
	keyPts := keyKun.Parent()
	// log.Infof(c, "No Cm adalah: %v", keyPts.StringID())
	err = datastore.Get(c, keyPts, &dat)
	if err != nil {
		LogError(c, err)
	}

	pts := &Pasien{
		NoCM:       keyPts.StringID(),
		NamaPasien: dat.NamaPasien,
		Diagnosis:  kun.Diagnosis,
		ATS:        kun.ATS,
		ShiftJaga:  kun.ShiftJaga,
		Dept:       kun.Bagian,
		IKI:        kun.GolIKI,
		LinkID:     link,
		TglAsli:    kun.JamDatangRiil.Add(time.Duration(8) * time.Hour),
	}

	return pts
}
