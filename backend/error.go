package backend

import (
	"google.golang.org/appengine/log"
	"golang.org/x/net/context"
)

func LogError(c context.Context, e error) {
	log.Errorf(c, "Error is: %v", e)
	return
}