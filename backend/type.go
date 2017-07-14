package backend

import (
	"time"
)

type MainView struct {
	Token  string   `json:"token"`
	User   string   `json:"user"`
	Bulan  []string `json:"bulan"`
	Pasien []Pasien `json:"pasien"`
	//IKI      []List    `json:"list"`
}

//Ini digunakan untuk view web, IKI1 dan IKI2 harus dipisah
type Pasien struct {
	StatusServer string    `json:"stat"`
	TglKunjungan string    `json:"tgl"`
	ShiftJaga    string    `json:"shift"`
	ATS          string    `json:"ats"`
	Bagian       string    `json:"bagian"`
	NoCM         string    `json:"nocm"`
	NamaPasien   string    `json:"nama"`
	Diagnosis    string    `json:"diag"`
	IKI          string    `json:"iki"`
	LinkID       string    `json:"link"`
	TglAsli      time.Time `json:"tglasli"`
}

//Ini untuk menyimpan jumlah iki yang diperoleh
type List struct {
	TglJaga string `json:"tgl"`
	//ShiftJaga    string `json:"shift"`
	SumIKI1 string `json:"iki1"`
	SumIKI2 string `json:"iki2"`
}

type KunjunganPasien struct {
	Diagnosis, LinkID      string
	GolIKI, ATS, ShiftJaga string
	JamDatang              time.Time
	Dokter                 string
	Hide                   bool
	JamDatangRiil          time.Time
	Bagian                 string
}

type DataPasien struct {
	NamaPasien, NomorCM, JenKel, Alamat string
	TglDaftar, Umur                     time.Time
}

type Kursor struct {
	Point string
}

type NavBar struct {
	Token string   `json:"token"`
	User  string   `json:"user"`
	Bulan []string `json:"bulan"`
}

type TokenError struct {
	ErrorTok string `json:"errtok"`
}

type InputPts struct {
	DataPasien      `json:"datapts"`
	KunjunganPasien `json:"kunjungan"`
}

type ServerResponse struct {
	Error  string `json:"error"`
	Pasien `json:"pasien"`
}
