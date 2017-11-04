package ats

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
type LembarATS struct {
	LinkID         string `json:"link"`
	KeluhanUtama   string `json:"kelut"`
	Subyektif      string `json:"subyektif"`
	TDSistolik     string `json:"tdsis"`
	TDDiastolik    string `json:"tddi"`
	Nadi           string `json:"nadi"`
	LajuPernafasan string `json:"rr"`
	SuhuBadan      string `json:"temp"`
	LokasiNyeri    string `json:"nyerilok"`
	NRS            string `json:"nrs"`
	Keterangan     string `json:"keterangan"`
	GCSE           string `json:"gcse"`
	GCSV           string `json:"gcsv"`
	GCSM           string `json:"gcsm"`
	TglInput       time.Time `json:"input"`
}
type RekamMedis struct {
    Pasien   Pasien    `json:"pasien"`
	LembarATS LembarATS `json:"lembarats"`
}

func init() {
	http.Handle("/simpan", CekToken(http.HandlerFunc(SimpanLembarATS)))
	http.Handle("/get-rm-kun", CekToken(http.HandlerFunc(GetRM)))
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

func SimpanLembarATS(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	ats := &LembarATS{}
	json.NewDecoder(r.Body).Decode(ats)
	defer r.Body.Close()
	keypar, err := datastore.DecodeKey(ats.LinkID)
	if err != nil {
		LogError(ctx, err)
	}
	key := datastore.NewIncompleteKey(ctx, "LembarATS", keypar)
	_, err = datastore.Put(ctx, key, ats)
	if err != nil {
		LogError(ctx, err)
	}
	json.NewEncoder(w).Encode(ats)
}
func ConvertDatastore(c context.Context, n KunjunganPasien, k *datastore.Key) *Pasien {
	tanggal := UbahTanggal(n.JamDatang, n.ShiftJaga)
	nocm, namapts := GetDataPts(c, k)

	m := &Pasien{
		TglKunjungan: tanggal,
		ShiftJaga:    n.ShiftJaga,
		NoCM:         nocm,
		NamaPasien:   namapts,
		Diagnosis:    n.Diagnosis,
		ATS:          n.ATS,
		Dept:         n.Bagian,
		IKI:          "",
		LinkID:       k.Encode(),
	}

	if n.GolIKI == "1" {
		m.IKI = "1"
	} else {
		m.IKI = ""
	}

	return m
}
func GetDataPts(c context.Context, k *datastore.Key) (no, nama string) {
	var p DataPasien
	keypar := k.Parent()
	err := datastore.Get(c, keypar, &p)
	if err != nil {
		log.Errorf(c, "error getting data: %v", err)
	}
	no = keypar.StringID()
	nama = p.NamaPasien
	return no, nama
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

func GetRM(w http.ResponseWriter, r *http.Request){
    ctx:= appengine.NewContext(r)
	ats := &LembarATS{}
	json.NewDecoder(r.Body).Decode(ats)
	defer r.Body.Close()
	keypar, err := datastore.DecodeKey(ats.LinkID)
	if err != nil {
		LogError(ctx, err)
	}
	kun := &KunjunganPasien{}
	err = datastore.Get(ctx, keypar, kun)
	if err != nil {
		LogError(ctx, err)
	}
	pts := ConvertDatastore(ctx, *kun, keypar)
	q := datastore.NewQuery("LembarATS").Ancestor(keypar)
	t := q.Run(ctx)
	for {
	    var at LembarATS
		_, err := t.Next(&at)
		if err == datastore.Done {
		    break
		}
		if err != nil {
		    LogError(ctx, err)
			break
		}
		rm := &RekamMedis{
		    Pasien: *pts,
			LembarATS: at,
		}
		json.NewEncoder(w).Encode(rm)
	}
}