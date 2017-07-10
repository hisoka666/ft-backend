package backend

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
)

func LogError(c context.Context, e error) {
	log.Errorf(c, "Error is: %v", e)
	return
}

// func PatientError(er string) *Pasien {
// 	pts := &Pasien{
// 		NoCM: er
// 	}

// 	return pts
// }

// func TokenError()
