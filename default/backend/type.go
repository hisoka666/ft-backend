package backend

import (
	"time"
)

type Staff struct {
	Email, NamaLengkap, LinkID, Peran string
}

// type MainView struct {
// 	Token  string   `json:"token"`
// 	User   string   `json:"user"`
// 	Bulan  []string `json:"bulan"`
// 	Pasien []Pasien `json:"pasien"`
// 	//IKI      []List    `json:"list"`
// }

//Ini digunakan untuk view web, IKI1 dan IKI2 harus dipisah
type Pasien struct {
	StatusServer string    `json:"stat"`
	TglKunjungan string    `json:"tgl"`
	ShiftJaga    string    `json:"shift"`
	ATS          string    `json:"ats"`
	Dept         string    `json:"dept"`
	NoCM         string    `json:"nocm"`
	NamaPasien   string    `json:"nama"`
	Diagnosis    string    `json:"diag"`
	IKI          string    `json:"iki"`
	LinkID       string    `json:"link"`
	TglAsli      time.Time `json:"tglasli"`
}
type SupervisorListPasien struct {
	TglKunjungan time.Time `json:"tgl"`
	ATS          string    `json:"ats"`
	Dept         string    `json:"dept"`
	Diagnosis    string    `json:"diag"`
	LinkID       string    `json:"link"`
}
type KursorIGD struct {
	Bulan string `json:"bulan"`
	Point string `json:"point"`
}
type SupervisorList struct {
	StatusServer    string                 `json:"status"`
	ListPasien      []SupervisorListPasien `json:"listpasien"`
	Token           string                 `json:"token"`
	SupervisorName  string                 `json:"user"`
	ListBulan       []string               `json:"listbulan"`
	PerHari         []int                  `json:"perhari"`
	PerDeptPerHari  []Departemen           `json:"perdept"`
	PerShiftPerHari []PerShift             `json:"shift"`
}

//Ini untuk menyimpan jumlah iki yang diperoleh
// type List struct {
// 	TglJaga string `json:"tgl"`
// 	//ShiftJaga    string `json:"shift"`
// 	SumIKI1 string `json:"iki1"`
// 	SumIKI2 string `json:"iki2"`
// }

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
	Point string `json:"point"`
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

type InputObat struct {
	MerkDagang     string   `json:"merk"`
	Kandungan      string   `json:"kand"`
	MinDose        string   `json:"mindose"`
	MaxDose        string   `json:"maxdose"`
	Tablet         []string `json:"tab"`
	Sirup          []string `json:"syr"`
	Drop           []string `json:"drop"`
	Lainnya        string   `json:"lainnya"`
	SediaanLainnya []string `json:"lainnya_sediaan"`
	Rekomendasi    string   `json:"rekom"`
	Dokter         string   `json:"doc"`
}

type Admin struct {
	Staff *[]Staff `json:"list"`
	Token string   `json:"token"`
}
type ListIKI struct {
	Tanggal int `json:"tgl"`
	SumIKI1 int `json:"iki1"`
	SumIKI2 int `json:"iki2"`
}
type MainView struct {
	Token      string         `json:"token"`
	User       string         `json:"user"`
	Bulan      []string       `json:"bulan"`
	Pasien     []Pasien       `json:"pasien"`
	IKI        []ListIKI      `json:"list"`
	Admin      Admin          `json:"admin"`
	Supervisor SupervisorList `json:"supervisor"`
	Peran      string         `json:"peran"`
}
type Departemen struct {
	Interna   int `json:"interna"`
	Bedah     int `json:"bedah"`
	Anak      int `json:"anak"`
	Obgyn     int `json:"obgyn"`
	Saraf     int `json:"saraf"`
	Anestesi  int `json:"anes"`
	Psikiatri int `json:"psik"`
	THT       int `json:"tht"`
	Kulit     int `json:"kulit"`
	Kardio    int `json:"jant"`
	Umum      int `json:"umum"`
	Mata      int `json:"mata"`
	MOD       int `json:"mod"`
}

type PerShift struct {
	Pagi  int `json:"pagi"`
	Sore  int `json:"sore"`
	Malam int `json:"malam"`
	Total int `json:"total"`
}
