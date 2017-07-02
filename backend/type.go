package backend

import(
	"time"
)

type WebView struct {
	Token    string    `json:"token"`
	User     string    `json:"user"`
	Pasien   []Pasien  `json:"pasien"`
	//IKI      []List    `json:"list"`
}

//Ini digunakan untuk view web, IKI1 dan IKI2 harus dipisah
type Pasien struct {
	TglKunjungan string `json:"tgl"`
	ShiftJaga    string `json:"shift"`
	NoCM         string `json:"nocm"`
	NamaPasien   string `json:"nama"`
	Diagnosis    string `json:"diag"`
	IKI1         string `json:"iki1"`
	IKI2         string `json:"iki2"`
	LinkID       string `json:"link"`
}

//Ini untuk menyimpan jumlah iki yang diperoleh
type List struct {
	TglJaga      string `json:"tgl"`
	//ShiftJaga    string `json:"shift"`
	SumIKI1         string `json:"iki1"`
	SumIKI2         string `json:"iki2"`
}

type KunjunganPasien struct {
	Diagnosis, LinkID      string
	GolIKI, ATS, ShiftJaga string
	JamDatang              time.Time
	Dokter                 string
	Hide                   bool
	JamDatangRiil          time.Time
}

type DataPasien struct {
	NamaPasien, NomorCM, JenKel, Alamat string
	TglDaftar, Umur                     time.Time
}