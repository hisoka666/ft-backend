package backend

import (
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)


func GetListbyKursor(c context.Context, email, tgl string) []Pasien {
	_, kurKey := DatastoreKey(c, "Dokter", email, "Kursor", tgl)
	log.Infof(c, "Key adalah: %s", kurKey)

	kur := Kursor{}
	kun := KunjunganPasien{}

	err := datastore.Get(c, kurKey, &kur)
	q := datastore.NewQuery("KunjunganPasien").Filter("Dokter=", email).Filter("Hide=", false).Order("-JamDatang")

	if err != nil && err != datastore.ErrNoSuchEntity {
		LogError(c, err)
	}

	if err != nil && err == datastore.ErrNoSuchEntity {
		t := q.Run(c)
		strmon := tgl + "/01"
		mon, err := time.Parse("2006/01/02", strmon)
		if err != nil {
			LogError(c, err)
		}
		for {
			_, err := t.Next(&kun)
			if err == datastore.Done {
				break
			}
			if err != nil {
				LogError(c, err)
			}

			jamEdit := AdjustTime(kun.JamDatang, kun.ShiftJaga)

			if jamEdit.After(mon) != true {
				cursor, _ := t.Cursor()
				kur.Point = cursor.String()
				if _, err := datastore.Put(c, kurKey, &kur); err != nil {
					LogError(c, err)
				}
				break
			}
		}

		err = datastore.Get(c, kurKey, &kur)
		if err != nil {
			LogError(c, err)
		}
	}

	kursor, err := datastore.DecodeCursor(kur.Point)
	if err != nil {
		LogError(c, err)
	}

	q = q.Start(kursor)

	var m []Pasien

	t := q.Run(c)
	strmonin := tgl + "/01"
	monin, err := time.Parse("2006/01/02", strmonin)
	log.Infof(c, "Tanggal satu adalah : %v", monin)

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
		log.Infof(c, "Jam datang adalah : %v", j.JamDatang)
		jamAdjust := AdjustTime(j.JamDatang, j.ShiftJaga)
		log.Infof(c, "Apakah sebelum? %v", jamAdjust.Before(monin))
		if jamAdjust.Before(monin) == true {
			break
		}

		n := ConvertDatastore(c, j, k)

		m = append(m, *n)
	}

	for i, j := 0, len(m)-1; i < j; i, j = i+1, j-1 {
		m[i], m[j] = m[j], m[i]
	}

	log.Infof(c, "list adalah : %v", m)
	return m
}
