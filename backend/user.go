package backend

import (
	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"
)

type Staff struct {
	Email, NamaLengkap, LinkID string
}

func CekStaff(ctx context.Context, email string) (user string, token string) {
	var staf []Staff
	q := datastore.NewQuery("Staff").Filter("Email=", email)

	_, err := q.GetAll(ctx, &staf)
	if err != nil {
		LogError(ctx, err)
	}
	user = ""
	token = ""
	if len(staf) == 0 {
		user = "no-access"
		return user, token
	}

	for _, v := range staf {
		token = CreateToken(ctx, v.Email)
		user = v.NamaLengkap
	}
	return user, token
}

func GetMainContent(c context.Context, user, token, email string) *WebView {
	web := &WebView{
		Token:  token,
		User:   user,
		Pasien: GetLast100(c, email),
	}
	return web
}

func UserKey(c context.Context, email string) *datastore.Key {
	gpKey := datastore.NewKey(c, "IGD", "fasttrack", 0, nil)
	parKey := datastore.NewKey(c, "Dokter", email, 0, gpKey)
	return parKey
}

func GetBulan(c context.Context, token, user string, k *datastore.Key) *NavBar {
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

    nav := &NavBar{
		Token: token,
		User: user,
		Bulan: list,
	}
	
	return nav

}

// func AppCtx(ctx appengine.Context, kind1 string, id1 string, kind2 string, id2 string) (string, *datastore.Key, *datastore.Key){

//    u := user.Current(ctx)
//    email := u.Email

//    gpKey := datastore.NewKey(ctx, "IGD", "fasttrack", 0, nil)
//    parKey := datastore.NewKey(ctx, kind1, id1, 0, gpKey)
//    chldKey := datastore.NewKey(ctx, kind2, id2, 0, parKey)

//    return email, parKey, chldKey
// }

// func ListLaporan(w http.ResponseWriter, r *http.Request) []string {
// 	ctx := appengine.NewContext(r)
// 	email, _, _ := AppCtx(ctx, "", "", "", "")
// 	_, key, _ := AppCtx(ctx, "Dokter", email, "Kursor", "")
// 	kur := []Kursor{}
// 	q := datastore.NewQuery("Kursor").Ancestor(key)
// 	keys, err := q.GetAll(ctx, &kur)
// 	if err != nil {
// 		fmt.Fprintln(w, "Error Fetching Kursor :", err)
// 	}

// 	var list []string
// 	for _, v := range keys {
// 		m := v.StringID()
// 		list = append(list, m)
// 	}
// 	for i, j := 0, len(list)-1; i < j; i, j = i+1, j-1 {
// 		list[i], list[j] = list[j], list[i]
// 	}
// 	return list
// }
