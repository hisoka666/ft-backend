package backend

import (
	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"
)

type Staff struct {
	Email, NamaLengkap, LinkID string
}

// func logError(c context.Context, e error) {
// 	log.Errorf(c, "Error is: %v", e)
// 	return
// }
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

func GetMainContent(c context.Context, user, token, email string) []byte {
	web := &WebView{
		Token:  token,
		User:   user,
		Pasien: GetLast100(c, email),
	}
	return ConvertJSON(web)
}
