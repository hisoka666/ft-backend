package update_entri

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
	http.Handle("/", CekToken(http.HandlerFunc(UpdateEntri)))
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

func UpdateEntri(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	ubah := &Pasien{}

	err := json.NewDecoder(r.Body).Decode(ubah)
	if err != nil {
		ubah.StatusServer = "kesalahan-decoding-json"
		LogError(ctx, err)
		json.NewEncoder(w).Encode(ubah)
	}

	up, err := getUpdateEntri(ctx, ubah)
	if err != nil {
		LogError(ctx, err)
		json.NewEncoder(w).Encode(up)
		return
	}

	json.NewEncoder(w).Encode(up)
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

func getUpdateEntri(c context.Context, n *Pasien) (*Pasien, error) {

	kun := &KunjunganPasien{}
	pts := &DataPasien{}

	keyKun, err := datastore.DecodeKey(n.LinkID)
	if err != nil {
		m := &Pasien{
			StatusServer: "kesalahan-decoding-Key",
		}
		return m, err
	}
	keyPts := keyKun.Parent()

	err = datastore.Get(c, keyKun, kun)
	if err != nil {
		m := &Pasien{
			StatusServer: "kesalahan-database-get-kunjungan-pts",
		}
		return m, err
	}

	kun.Diagnosis = n.Diagnosis
	kun.ATS = n.ATS
	kun.GolIKI = n.IKI
	kun.ShiftJaga = n.ShiftJaga
	kun.Bagian = n.Dept
	n.TglKunjungan = UbahTanggal(kun.JamDatang, kun.ShiftJaga)

	err = datastore.Get(c, keyPts, pts)
	if err != nil {
		m := &Pasien{
			NoCM: "kesalahan-database-get-datapts",
		}
		return m, err
	}
	pts.NamaPasien = n.NamaPasien

	if _, err := datastore.Put(c, keyKun, kun); err != nil {
		m := &Pasien{
			StatusServer: "kesalahan-database-put-kunjungan-failed",
		}
		return m, err
	}

	if _, err := datastore.Put(c, keyPts, pts); err != nil {
		m := &Pasien{
			StatusServer: "kesalahan-database-put-datapts-failed",
		}
		return m, err
	}

	if kun.GolIKI == "1" {
		n.IKI = "1"
	} else {
		n.IKI = ""
	}
	n.NoCM = keyPts.StringID()
	n.StatusServer = "OK"
	log.Infof(c, string(ConvertJSON(n)))
	return n, nil
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

func ConvertJSON(n interface{}) []byte {
	m, _ := json.Marshal(n)
	return m
}

func UbahTanggal(t time.Time, s string) string {
	zone, _ := time.LoadLocation("Asia/Makassar")
	idn := t.In(zone)
	jam := idn.Hour()
	tglstring := ""
	if jam < 12 && s == "3" {
		tglstring = idn.AddDate(0, 0, -1).Format("02-01-2006")
	} else {
		tglstring = idn.Format("02-01-2006")
	}
	return tglstring + " (" + StringShift(s) + ")"
}

func StringShift(n string) string {
	var m string
	switch n {
	case "1":
		m = "Pagi"
	case "2":
		m = "Sore"
	case "3":
		m = "Malam"
	}
	return m
}
