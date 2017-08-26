package backend

import (
	"golang.org/x/net/context"
	
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

func SupervisorPage(c context.Context){
	skrng, zone := CreateTime()
	awal := time.Date(skrng.Year(), skrng.Month(), 1, 8, 0, 0, 0, zone)
	q := datastore.NewQuery("KunjunganPasien").Filter("JamDatang >=", awal).Order("-JamDatang")
	t := q.Run(c)
	kun := []KunjunganPasien{}
	dat := KunjunganPasien{}
	for {

	}
}

func CreateTime() time.Time, *Location {
	t := time.Now()
	zone, err := time.LoadLocation("Asia/Makassar")
	if err != nil {
		fmt.Println("Err: ", err.Error())
	}
	jam := t.In(zone)
	return jam, zone
}