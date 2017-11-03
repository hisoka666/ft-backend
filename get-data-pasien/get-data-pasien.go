package get_data_pasien

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"cloud.google.com/go/storage"
	jwt "github.com/dgrijalva/jwt-go"

	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

func init() {
	http.Handle("/", CekToken(http.HandlerFunc(getCM)))
	http.Handle("/get-detail", CekToken(http.HandlerFunc(getDetail)))
	http.Handle("/add-surat-sakit", CekToken(http.HandlerFunc(addSuratSakit)))
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

type DataPasien struct {
	NamaPasien string    `json:"namapts"`
	NomorCM    string    `json:"nocm"`
	JenKel     string    `json:"jenkel"`
	Alamat     string    `json:"alamat"`
	TglDaftar  time.Time `json:"tgldaf"`
	TglLahir   time.Time `json:"tgllhr"`
	Umur       time.Time `json:"umur"`
}
type SuratSakit struct {
	LinkID        string `json:"link"`
	TglLahir      string `json:"tgl"`
	Pekerjaan     string `json:"pekerjaan"`
	Alamat        string `json:"alamat"`
	LamaIstirahat string `json:"lama"`
	StatusData    string `json:"status"`
	Dokter        string `json:"dokter"`
}

type DataSuratSakit struct {
	TglSurat      time.Time `json:"tglsurat"`
	NomorSurat    int       `json:"nomor"`
	LamaIstirahat string    `json:"lama"`
	Pekerjaan     string    `json:"pekerjaan"`
	LinkSurat     string    `json:"link"`
	Dokter        string    `json:"dokter"`
	NamaPasien    string    `json:"namapts"`
	NoCM          string    `json:"nocm"`
	Umur          string    `json:"umur"`
	Alamat        string    `json:"alamat"`
}

func addSuratSakit(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	zone, _ := time.LoadLocation("Asia/Makassar")
	sur := &SuratSakit{}
	bulskrng := time.Now().In(zone).Format("01/2006")
	kindSuratSakit := "DataSuratSakit" + bulskrng
	json.NewDecoder(r.Body).Decode(sur)
	// log.Infof(ctx, "Link adalah: %v", sur.LinkID)
	// log.Infof(ctx, "Data status adalah: %v", sur.StatusData)
	lahir, err := time.ParseInLocation("02-01-2006", sur.TglLahir, zone)
	if err != nil {
		log.Errorf(ctx, "Gagal parsing tanggal: %v", err)
	}
	umur := countAge(lahir)
	keyEntri, err := datastore.DecodeKey(sur.LinkID)
	if err != nil {
		log.Errorf(ctx, "Gagal mengubah kunci: %v", err)
	}
	keyPar := keyEntri.Parent()
	datapts := &DataPasien{}
	err = datastore.Get(ctx, keyPar, datapts)
	if err != nil {
		log.Errorf(ctx, "Gagal Mengambil data: %v", err)
	}
	keySurat := datastore.NewIncompleteKey(ctx, kindSuratSakit, keyEntri)
	q := datastore.NewQuery(kindSuratSakit).Order("-TglSurat")
	t := q.Run(ctx)
	for {
		j := &DataSuratSakit{}
		_, err := t.Next(j)
		if err == datastore.ErrNoSuchEntity || err == datastore.Done {
			if sur.StatusData == "data-ada" {
				data := &DataSuratSakit{
					TglSurat:      time.Now().In(zone),
					NomorSurat:    1,
					LamaIstirahat: sur.LamaIstirahat,
					Pekerjaan:     sur.Pekerjaan,
					LinkSurat:     keySurat.Encode(),
					Umur:          umur,
					NamaPasien:    datapts.NamaPasien,
					Dokter:        sur.Dokter,
					NoCM:          keyPar.StringID(),
					Alamat:        sur.Alamat,
				}
				_, err := datastore.Put(ctx, keySurat, data)
				if err != nil {
					log.Infof(ctx, "Gagal menyimpan surat: %v", err)
					return
				}
				data.NoCM = keyPar.StringID()
				json.NewEncoder(w).Encode(data)
			} else {
				datapts.TglLahir = lahir
				datapts.Alamat = sur.Alamat
				_, err = datastore.Put(ctx, keyPar, datapts)
				if err != nil {
					log.Errorf(ctx, "Gagal menyimpan data: %v", err)
				}
				data := &DataSuratSakit{
					TglSurat:      time.Now().In(zone),
					NomorSurat:    1,
					LamaIstirahat: sur.LamaIstirahat,
					Pekerjaan:     sur.Pekerjaan,
					LinkSurat:     keySurat.Encode(),
					Umur:          umur,
					NamaPasien:    datapts.NamaPasien,
					Dokter:        sur.Dokter,
					NoCM:          keyPar.StringID(),
					Alamat:        sur.Alamat,
				}
				_, err = datastore.Put(ctx, keySurat, data)
				if err != nil {
					log.Infof(ctx, "Gagal menyimpan surat: %v", err)
					return
				}
				data.NoCM = keyPar.StringID()
				json.NewEncoder(w).Encode(data)

			}
			log.Infof(ctx, "berhasil menyimpan surat")
			break
		}

		if sur.StatusData == "data-ada" {
			m := &DataSuratSakit{
				TglSurat:      time.Now().In(zone),
				NomorSurat:    j.NomorSurat + 1,
				LamaIstirahat: sur.LamaIstirahat,
				Pekerjaan:     sur.Pekerjaan,
				LinkSurat:     keySurat.Encode(),
				Umur:          umur,
				NamaPasien:    datapts.NamaPasien,
				Dokter:        sur.Dokter,
				NoCM:          keyPar.StringID(),
				Alamat:        sur.Alamat,
			}
			_, err := datastore.Put(ctx, keySurat, m)
			if err != nil {
				log.Infof(ctx, "Gagal menyimpan surat: %v", err)
				return
			}
			m.NoCM = keyPar.StringID()
			json.NewEncoder(w).Encode(m)
		} else {
			datapts.TglLahir = lahir
			datapts.Alamat = sur.Alamat
			_, err = datastore.Put(ctx, keyPar, datapts)
			if err != nil {
				log.Errorf(ctx, "Gagal menyimpan data: %v", err)
			}
			data := &DataSuratSakit{
				TglSurat:      time.Now().In(zone),
				NomorSurat:    j.NomorSurat + 1,
				LamaIstirahat: sur.LamaIstirahat,
				Pekerjaan:     sur.Pekerjaan,
				LinkSurat:     keySurat.Encode(),
				Umur:          umur,
				NamaPasien:    datapts.NamaPasien,
				Dokter:        sur.Dokter,
				NoCM:          keyPar.StringID(),
				Alamat:        sur.Alamat,
			}
			_, err = datastore.Put(ctx, keySurat, data)
			if err != nil {
				log.Infof(ctx, "Gagal menyimpan surat: %v", err)
				return
			}
			data.NoCM = keyPar.StringID()
			json.NewEncoder(w).Encode(data)
		}
		log.Infof(ctx, "berhasil menyimpan surat")
		break
	}

}
func countAge(birth time.Time) string {
	zone, _ := time.LoadLocation("Asia/Makassar")
	today := time.Now().In(zone)
	age := today.Year() - birth.Year()
	if today.YearDay() < birth.YearDay() {
		age--
	}
	strAge := strconv.Itoa(age)
	return strAge
}

func getDetail(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	pts := &Pasien{}

	err := json.NewDecoder(r.Body).Decode(pts)
	if err != nil {
		LogError(ctx, err)
	}
	key, err := datastore.DecodeKey(pts.LinkID)
	dat := &DataPasien{}
	keyPar := key.Parent()
	log.Infof(ctx, "key parent adalah : %v", keyPar)
	err = datastore.Get(ctx, keyPar, dat)
	if err != nil {
		LogError(ctx, err)
	}
	json.NewEncoder(w).Encode(dat)
}
func getCM(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	pts := &Pasien{}

	if r.Body == nil {
		pts.NamaPasien = "request-body-empty"
		json.NewEncoder(w).Encode(pts)
	}

	err := json.NewDecoder(r.Body).Decode(&pts)
	if err != nil {
		pts.NamaPasien = "kesalahan-decoding-json"
		json.NewEncoder(w).Encode(pts)
	}

	dat, err := GetNamaByNoCM(ctx, pts.NoCM)
	if err != nil {
		dat.NomorCM = "kesalahan-server"
		json.NewEncoder(w).Encode(pts)
	}
	pts.NamaPasien = dat.NamaPasien
	pts.NoCM = dat.NomorCM
	json.NewEncoder(w).Encode(pts)

}

func GetNamaByNoCM(c context.Context, nocm string) (DataPasien, error) {
	var pts DataPasien
	parKey := datastore.NewKey(c, "IGD", "fasttrack", 0, nil)
	ptsKey := datastore.NewKey(c, "DataPasien", nocm, 0, parKey)

	err := datastore.Get(c, ptsKey, &pts)
	if err != nil && err != datastore.ErrNoSuchEntity {
		return pts, err
	}

	return pts, nil
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
