package ft

import (
	"encoding/json"
	"fmt"
	"ft"
	"io/ioutil"
	"net/http"
	"time"

	"google.golang.org/appengine"
	//urlfetch digunakan untuk mengganti http.Get dan http.Post karena tidak didukung oleh app engine
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
	"google.golang.org/appengine/user"
)

type Staff struct {
	Email       string
	NamaLengkap string
	LinkID      string
}

type EntriPasien struct {
	Tgl       string `json:"tgl"`
	Jaga      string `json:"shift"`
	NoCM      string `json:"nocm"`
	NamaPts   string `json:"nama"`
	Diagnosis string `json:"diag"`
	IKI       string `json:"iki"`
	Entri     string `json:"id"`
}

type TokenResp struct {
	Token []string `json:"token"`
	//Ent []EntriPasien `json:"list"`
}

type EntriList struct {
	List []EntriPasien `json:"list"`
}

type EntriResp struct {
	Ent *EntriList `json:"ent"`
}
type Resp struct {
	Resp *TokenResp `json:"resp"`
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

func init() {
	http.Handle("/testuser", ft.CekToken(http.HandlerFunc(hello)))
	http.HandleFunc("/test", test)
}

func test(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	//http.Get versi appengine
	client := urlfetch.Client(ctx)
	token := "https://www.googleapis.com/oauth2/v3/tokeninfo?id_token=" + r.FormValue("token")
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
		log.Infof(ctx, origin)
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
			w.Header().Set("Content-Type", "application/json")
			reply := CreateJSON("no-access")
			w.Write(reply)
			return
		}

		//Setelah ketemu, membuat respon untuk Ajax
		for _, v := range staf {
			tok := CreateJSON(ft.CreateToken(w, r, v.Email))
			log.Infof(ctx, string(tok))
			w.Header().Set("Content-Type", "application/json")
			w.Write(tok)

		}
	}
}

func CreateJSON(token string) []byte {
	m := &TokenResp{
		Token: []string{token},
	}

	r := &Resp{
		Resp: m,
	}

	res, _ := json.Marshal(r)

	return res
}
func hello(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	u := user.Current(ctx)
	log.Infof(ctx, u.Email)
	result := Last100(r, u.Email)

	w.Header().Set("Content-Type", "application/json")
	w.Write(CreateJSONList(result))
}
func CreateJSONList(n []EntriPasien) []byte {
	m := &EntriList{
		List: n,
	}

	r := &EntriResp{
		Ent: m,
	}

	res, _ := json.Marshal(r)

	return res

}
func Last100(r *http.Request, email string) []EntriPasien {
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

	// jsm, err := json.Marshal(m)
	// if err != nil {
	// 	log.Errorf(ctx, "error marshalling json: %v", err)
	// }
	return m
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
