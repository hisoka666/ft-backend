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
	http.Handle("/", CekToken(http.HandlerFunc(addObatPasien)))
}

// type DetailPasien struct {
// 	Pasien    DataPasien        `json:"datapts"`
// 	Kunjungan []KunjunganPasien `json:"kunjungan"`
// 	LinkID    string            `json:"link"`
// }

// type KunjunganPasien struct {
// 	Diagnosis, LinkID      string
// 	GolIKI, ATS, ShiftJaga string
// 	JamDatang              time.Time
// 	Dokter                 string
// 	Hide                   bool
// 	JamDatangRiil          time.Time
// 	Bagian                 string
// }
// type DataPasien struct {
// 	NamaPasien string    `json:"namapts"`
// 	NomorCM    string    `json:"nocm"`
// 	JenKel     string    `json:"jenkel"`
// 	Alamat     string    `json:"alamat"`
// 	TglDaftar  time.Time `json:"tgldaf"`
// 	TglLahir   time.Time `json:"tgllhr"`
// 	Umur       time.Time `json:"umur"`
// }

// type Pasien struct {
// 	StatusServer string    `json:"stat"`
// 	TglKunjungan string    `json:"tgl"`
// 	ShiftJaga    string    `json:"shift"`
// 	ATS          string    `json:"ats"`
// 	Dept         string    `json:"dept"`
// 	NoCM         string    `json:"nocm"`
// 	NamaPasien   string    `json:"nama"`
// 	Diagnosis    string    `json:"diag"`
// 	IKI          string    `json:"iki"`
// 	LinkID       string    `json:"link"`
// 	TglAsli      time.Time `json:"tglasli"`
// }
type PasienResep struct {
	Nama      string `json:"nama"`
	Umur      string `json:"umur"`
	Berat     string `json:"berat"`
	Alamat    string `json:"alamat"`
	Alergi    string `json:"alergi"`
	Diagnosis string `json:"diag"`
	NoCM      string `json:"nocm"`
	LinkID    string `json:"link"`
}
type Resep struct {
	Dokter    string      `json:"dokter"`
	Tanggal   string      `json:"tanggal"`
	ListObat  []Obat      `json:"listobat"`
	ListPuyer []Puyer     `json:"listpuyer"`
	Pasien    PasienResep `json:"pasien"`
}

type Obat struct {
	NamaObat   string `json:"obat"`
	Jumlah     string `json:"jumlah"`
	Instruksi  string `json:"instruksi"`
	Keterangan string `json:"keterangan"`
}

type Puyer struct {
	Obat       []SatuObat `json:"satuobat"`
	Racikan    string     `json:"racikan"`
	JmlRacikan string     `json:"jml-racikan"`
	Instruksi  string     `json:"instruksi"`
	Keterangan string     `json:"keterangan"`
}

type SatuObat struct {
	NamaObat string `json:"obat"`
	Takaran  string `json:"takaran"`
}
type JSONResep struct {
	Tanggal    time.Time `json:"tanggal"`
	JSONString string    `json:"jsonstring"`
}

func addObatPasien(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	obt := &Resep{}
	json.NewDecoder(r.Body).Decode(obt)
	key, err := datastore.DecodeKey(obt.Pasien.LinkID)
	if err != nil {
		LogError(ctx, err)
	}
	js, err := json.Marshal(obt)
	if err != nil {
		LogError(ctx, err)
	}
	zone, _ := time.LoadLocation("Asia/Makassar")
	jos := &JSONResep{
		Tanggal:    time.Now().In(zone),
		JSONString: string(js),
	}
	// log.Infof(ctx, "key adalah: %v", key)
	keyObt := datastore.NewKey(ctx, "JSONResep", "", 0, key)
	// log.Infof(ctx, "keyobat adalah: %v", keyObt)
	_, err = datastore.Put(ctx, keyObt, jos)
	if err != nil {
		LogError(ctx, err)
	}
	resp := &SatuObat{
		NamaObat: "OK",
	}

	json.NewEncoder(w).Encode(resp)
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
