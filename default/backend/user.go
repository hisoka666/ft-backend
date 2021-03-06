package backend

import (
	"time"

	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

func CekStaff(ctx context.Context, email string) (user, token, peran, link string) {
	var staf []Staff
	q := datastore.NewQuery("Staff").Filter("Email=", email)

	k, err := q.GetAll(ctx, &staf)
	if err != nil {
		LogError(ctx, err)
	}
	user = ""
	token = ""
	peran = ""
	link = ""
	if len(staf) == 0 {
		user = "no-access"
		return user, token, peran, link
	}

	for _, v := range staf {
		token = CreateToken(ctx, v.Email)
		user = v.NamaLengkap
		peran = v.Peran
		link = k[0].Encode()
	}
	// var kunci *datastore.Key
	// for _, n := range k {
	// 	kunci = n
	// }
	log.Infof(ctx, "Link adalah: %v", link)
	return user, token, peran, link
}

func GetMainContent(c context.Context, user, token, link, peran, email string) *MainView {
	web := &MainView{
		Token:  token,
		User:   user,
		Bulan:  GetBulan(c, UserKey(c, email)),
		Pasien: GetLast100(c, email),
		LinkID: link,
		Peran:  peran,
	}
	log.Infof(c, "Link dari get main adalah: %v", web.LinkID)
	return web
}

func UserKey(c context.Context, email string) *datastore.Key {
	gpKey := datastore.NewKey(c, "IGD", "fasttrack", 0, nil)
	parKey := datastore.NewKey(c, "Dokter", email, 0, gpKey)
	return parKey
}

func GetBulan(c context.Context, k *datastore.Key) []string {
	kur := []Kursor{}

	q := datastore.NewQuery("Kursor").Ancestor(k)
	n, err := q.GetAll(c, &kur)
	if err != nil {
		LogError(c, err)
	}

	var list []string

	for _, v := range n {
		m := v.StringID()
		list = append(list, m)
	}

	for i, j := 0, len(list)-1; i < j; i, j = i+1, j-1 {
		list[i], list[j] = list[j], list[i]
	}

	return list

}

func LogError(c context.Context, e error) {
	log.Errorf(c, "Error is: %v", e)
	return
}

func GetLast100(c context.Context, email string) []Pasien {
	q := datastore.NewQuery("KunjunganPasien").Filter("Dokter =", email).Filter("Hide =", false).Order("-JamDatang").Limit(100)
	var m []Pasien

	t := q.Run(c)

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

		n := ConvertDatastore(c, j, k)

		m = append(m, *n)

	}
	return m

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
