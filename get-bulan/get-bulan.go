package get_bulan

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

type MainView struct {
	Token  string   `json:"token"`
	User   string   `json:"user"`
	Bulan  []string `json:"bulan"`
	Pasien []Pasien `json:"pasien"`
	//IKI      []List    `json:"list"`
}
type Kursor struct {
	Point string `json:"point"`
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

func init() {
	http.Handle("/", CekToken(http.HandlerFunc(getBulan)))
	http.Handle("/bulanini", CekToken(http.HandlerFunc(getBulanIni)))
}

func getBulanIni(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	pts := MainView{}
	json.NewDecoder(r.Body).Decode(&pts)
	tgl := pts.Bulan[0]
	log.Infof(ctx, "User %s sedang mencoba mengakses data %s", pts.User, tgl)
	nn := GetBulanIniList(ctx, pts.User, tgl)
	log.Infof(ctx, "List bulan ini adalah: %v", nn)
	json.NewEncoder(w).Encode(nn)
}

func getBulan(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	pts := MainView{}
	json.NewDecoder(r.Body).Decode(&pts)
	tgl := pts.Bulan[0]
	log.Infof(ctx, "User %s sedang mencoba mengakses data %s", pts.User, tgl)
	nn := GetListbyKursor(ctx, pts.User, tgl)
	json.NewEncoder(w).Encode(nn)
}
func GetBulanIniList(c context.Context, email, tgl string) []Pasien {
	q := datastore.NewQuery("KunjunganPasien").Filter("Dokter=", email).Filter("Hide=", false).Order("-JamDatang")
	// strmon := tgl + "/01 08:00:00"
	strmon := tgl + "/01 00:00:00 +0800"
	// zone, _ := time.LoadLocation("Asia/Makassar")
	// mon, err := time.Parse("2006/01/02 15:04:05", strmon)
	mon, err := time.Parse("2006/01/02 15:04:05 -0700", strmon)
	if err != nil {
		LogError(c, err)
	}
	log.Infof(c, "Awal bulan adalah: %v", mon)
	t := q.Run(c)
	if err != nil {
		LogError(c, err)
	}
	var kun KunjunganPasien
	var m []Pasien
	for {
		k, err := t.Next(&kun)
		if err == datastore.Done {
			break
		}
		if err != nil {
			LogError(c, err)
		}

		jamEdit := AdjustTime(kun.JamDatang, kun.ShiftJaga)
		log.Infof(c, "Jam adjusted adalah: %v", jamEdit)
		log.Infof(c, "jam before true? %v", jamEdit.Before(mon))
		if jamEdit.Before(mon) == true {
			break
		}

		n := ConvertDatastore(c, kun, k)
		log.Infof(c, "Converted database: %v", n)
		m = append(m, *n)
	}
	for i, j := 0, len(m)-1; i < j; i, j = i+1, j-1 {
		m[i], m[j] = m[j], m[i]
	}

	return m
}
func GetListbyKursor(c context.Context, email, tgl string) []Pasien {
	_, kurKey := DatastoreKey(c, "Dokter", email, "Kursor", tgl)
	log.Infof(c, "Key adalah: %s", kurKey)
	yr, _ := strconv.Atoi(tgl[:4])
	mo, _ := strconv.Atoi(tgl[5:7])
	zone, _ := time.LoadLocation("Asia/Makassar")
	q := datastore.NewQuery("KunjunganPasien").Filter("Dokter=", email).Filter("Hide=", false).Order("-JamDatang")

	kur := Kursor{}
	err := datastore.Get(c, kurKey, &kur)
	if err != nil {
		LogError(c, err)
	}
	log.Infof(c, "Tanggal adalah : %v", tgl)

	kursor, err := datastore.DecodeCursor(kur.Point)
	if err != nil {
		LogError(c, err)
	}

	q = q.Start(kursor)

	var m []Pasien

	t := q.Run(c)
	monin := time.Date(yr, time.Month(mo), 1, 0, 0, 0, 0, zone)

	for {
		var j KunjunganPasien
		k, err := t.Next(&j)
		if err == datastore.Done {
			break
		}

		if err != nil {
			LogError(c, err)
			break
		}
		// log.Infof(c, "Jam datang adalah : %v", j.JamDatang)
		jamAdjust := AdjustTime(j.JamDatang, j.ShiftJaga)
		// log.Infof(c, "Apakah sebelum? %v", jamAdjust.Before(monin))
		if jamAdjust.Before(monin) == true {
			break
		}

		n := ConvertDatastore(c, j, k)
		log.Infof(c, "Tanggal ini isinya: %v", n)

		m = append(m, *n)
	}

	for i, j := 0, len(m)-1; i < j; i, j = i+1, j-1 {
		m[i], m[j] = m[j], m[i]
	}

	log.Infof(c, "list adalah : %v", m)
	return m
}
func AdjustTime(t time.Time, s string) time.Time {
	zone, _ := time.LoadLocation("Asia/Makassar")
	idn := t.In(zone)
	jam := idn.Hour()
	// tglstring := ""
	if jam < 12 && s == "3" {
		// tglstring =
		return idn.AddDate(0, 0, -1)
	} else {
		// tglstring =
		return idn
	}
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
