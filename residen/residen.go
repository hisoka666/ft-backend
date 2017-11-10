package residen

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
	http.Handle("/get-new-pasien", CekToken(http.HandlerFunc(getNewPasien)))
	http.Handle("/get-refresh", CekToken(http.HandlerFunc(getRefresh)))
}

type ListPasienResiden struct {
	List []Pasien `json:"listpasien"`
}
type Staff struct {
	Email, NamaLengkap, LinkID, Peran string
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
func ConvertDatastore(c context.Context, n KunjunganPasien, k *datastore.Key) *Pasien {
	tanggal := UbahTanggal(n.JamDatang, n.ShiftJaga)
	nocm, namapts := GetDataPts(c, k)
	zone, _ := time.LoadLocation("Asia/Makassar")

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
		TglAsli:      n.JamDatang.In(zone),
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
func getNewPasien(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	kun := &KunjunganPasien{}
	json.NewDecoder(r.Body).Decode(kun)
	defer r.Body.Close()
	log.Infof(ctx, "Jam datang kiriman dari client: %v", kun.JamDatang)
	st := datastore.NewQuery("Staff").Filter("Email =", kun.Dokter)
	dept := ""
	f := &dept
	s := st.Run(ctx)
	for {
		n := &Staff{}
		_, err := s.Next(n)
		if err != nil {
			break
		}
		*f = n.Peran[8:]
	}
	log.Infof(ctx, "Peran dari client : %v", dept)
	kn := []Pasien{}
	q := datastore.NewQuery("KunjunganPasien").Order("-JamDatang").Filter("Hide =", false)
	t := q.Run(ctx)
	for {
		j := &KunjunganPasien{}
		k, err := t.Next(j)
		if err != nil {
			break
		}
		log.Infof(ctx, "Bagian pasien adalah: %v", j.Bagian)
		if j.Bagian != dept {
			continue
		}
		p := ConvertDatastore(ctx, *j, k)
		log.Infof(ctx, "Jam datang pasien adalah: %v", p.TglAsli)
		if p.TglAsli.Before(kun.JamDatang) {
			break
		}
		log.Errorf(ctx, "data pasien baru masuk: %v", p.NamaPasien)

		kn = append(kn, *p)
	}
	// log.Infof(ctx, "banyak pasien baru adalah: %v", len(kn))
	// for _, m := range kn {
	// 	log.Infof(ctx, "nama pasien adalah: %v", m.NamaPasien)
	// }
	list := &ListPasienResiden{
		List: kn,
	}
	json.NewEncoder(w).Encode(list)
}

func getRefresh(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	kun := &KunjunganPasien{}
	json.NewDecoder(r.Body).Decode(kun)
	defer r.Body.Close()
	log.Infof(ctx, "Jam datang kiriman dari client: %v", kun.JamDatang)
	st := datastore.NewQuery("Staff").Filter("Email =", kun.Dokter)
	dept := ""
	f := &dept
	s := st.Run(ctx)
	for {
		n := &Staff{}
		_, err := s.Next(n)
		if err != nil {
			break
		}
		*f = n.Peran[8:]
	}
	kn := []Pasien{}
	q := datastore.NewQuery("KunjunganPasien").Order("-JamDatang").Filter("Hide =", false)
	t := q.Run(ctx)
	for {
		j := &KunjunganPasien{}
		k, err := t.Next(j)
		if err != nil {
			break
		}
		log.Infof(ctx, "Bagian pasien adalah: %v", j.Bagian)
		if j.Bagian != dept {
			continue
		}
		p := ConvertDatastore(ctx, *j, k)
		kn = append(kn, *p)
		if len(kn) >= 100 {
			break
		}
	}

	list := &ListPasienResiden{
		List: kn,
	}
	json.NewEncoder(w).Encode(list)

}
