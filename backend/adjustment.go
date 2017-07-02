package backend

import (
	"time"
)

func UbahTanggal(t time.Time, s string) string {
	jam := t.Hour()
	tglstring := ""
	if jam < 12 && s == "3" {
		tglstring = t.AddDate(0, 0, -1).Format("02-01-2006")
	} else {
		tglstring = t.Format("02-01-2006")
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

func DatebyInt(m, y int) time.Time {
	in := time.Month(m)
	monIn := time.Date(y, in, 1, 0, 0, 0, 0, time.UTC)

	return monIn
}
