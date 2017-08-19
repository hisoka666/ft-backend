package add_pasien

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"google.golang.org/appengine/search"

	"cloud.google.com/go/storage"
	jwt "github.com/dgrijalva/jwt-go"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

func init() {
	http.Handle("/", CekToken(http.HandlerFunc(inputPasien)))
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
type InputPts struct {
	DataPasien      `json:"datapts"`
	KunjunganPasien `json:"kunjungan"`
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
type IndexDataPasien struct {
	Nama string
	NoCM string
}

func inputPasien(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	input := &InputPts{}
	pts := &Pasien{}
	err := json.NewDecoder(r.Body).Decode(input)
	if err != nil {
		pts.NoCM = "kesalahan-decoding-json"
		json.NewEncoder(w).Encode(input)
	}

	data := &DataPasien{
		NamaPasien: input.DataPasien.NamaPasien,
		TglDaftar:  input.DataPasien.TglDaftar,
	}

	kun := &KunjunganPasien{
		Diagnosis:     input.KunjunganPasien.Diagnosis,
		GolIKI:        input.KunjunganPasien.GolIKI,
		ATS:           input.KunjunganPasien.ATS,
		ShiftJaga:     input.KunjunganPasien.ShiftJaga,
		JamDatang:     input.KunjunganPasien.JamDatang,
		JamDatangRiil: input.KunjunganPasien.JamDatangRiil,
		Dokter:        input.KunjunganPasien.Dokter,
		Bagian:        input.KunjunganPasien.Bagian,
	}

	numKey, inputKey := DatastoreKey(ctx, "DataPasien", input.DataPasien.NomorCM, "KunjunganPasien", "")

	if input.DataPasien.TglDaftar.IsZero() == false {
		k, err := datastore.Put(ctx, numKey, data)
		if err != nil {
			LogError(ctx, err)
			pts.NoCM = "kesalahan-database"
			json.NewEncoder(w).Encode(input)
			return
		}

		ind := &IndexDataPasien{
			Nama: input.DataPasien.NamaPasien,
			NoCM: input.DataPasien.NomorCM,
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

	k, err := datastore.Put(ctx, inputKey, kun)
	if err != nil {
		LogError(ctx, err)
		pts.NoCM = "kesalahan-database"
		json.NewEncoder(w).Encode(pts)
		return
	}

	pts = ConvertDatastore(ctx, input.KunjunganPasien, k)

	json.NewEncoder(w).Encode(pts)

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

func DatastoreKey(ctx context.Context, kind1 string, id1 string, kind2 string, id2 string) (*datastore.Key, *datastore.Key) {
	gpKey := datastore.NewKey(ctx, "IGD", "fasttrack", 0, nil)
	parKey := datastore.NewKey(ctx, kind1, id1, 0, gpKey)
	chldKey := datastore.NewKey(ctx, kind2, id2, 0, parKey)

	return parKey, chldKey
}
