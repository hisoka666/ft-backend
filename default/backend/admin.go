package backend

import (
	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

func AdminPage(c context.Context, token string) MainView {
	staf := []Staff{}
	st := &Staff{}
	q := datastore.NewQuery("Staff").Filter("Peran =", "staf")
	p := datastore.NewQuery("Staff").Filter("Peran =", "supervisor")
	// q := datastore.NewQuery("Staff")
	s := p.Run(c)
	for {
		k, err := s.Next(st)
		if err == datastore.Done {
			break
		}
		if err != nil {
			LogError(c, err)
		}
		st.LinkID = k.Encode()
		staf = append(staf, *st)
	}
	t := q.Run(c)
	for {
		k, err := t.Next(st)
		if err == datastore.Done {
			break
		}
		if err != nil {
			LogError(c, err)
		}

		st.LinkID = k.Encode()
		staf = append(staf, *st)
	}

	// _, err := q.GetAll(c, &staf)
	// if err != nil {
	// 	LogError(c, err)
	// }
	log.Infof(c, "List staf adalah: %v", staf)
	admin := Admin{
		Staff: &staf,
		Token: token,
	}
	// log.Infof(c, "List staf adalah: %v", admin)
	main := MainView{
		Token: token,
		User:  "you-know-who",
		Admin: admin,
		Peran: "admin",
	}
	return main
}
