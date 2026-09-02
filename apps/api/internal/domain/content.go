package domain

import "context"

// The site's content tables are all the same shape of problem: an ordered list,
// usually filtered to the active rows. They are modelled as separate types
// rather than one generic "content item" because their columns genuinely
// differ, and collapsing them would mean every consumer unpacking a bag of
// optional fields.

type Regulation struct {
	ID     int64
	Jenis  string
	Nomor  string
	Tahun  int
	Judul  string
	File   string
	Urutan int
}

type PublicApp struct {
	ID        int64
	Nama      string
	Deskripsi string
	URL       string
	Icon      string
	Target    string
	Urutan    int
}

type Video struct {
	ID        int64
	Judul     string
	Deskripsi string
	Thumbnail string
	URLVideo  string
	Status    string
	Tanggal   string
}

// PerformanceDoc is a quarterly investment report (the `twdata` table).
type PerformanceDoc struct {
	ID       int64
	TWOption string
	FilePath string
}

type ServiceLocation struct {
	ID        int64
	Judul     string
	Deskripsi string
	Gambar    string
	AltText   string
	Lokasi    string
	Alamat    string
	Warna     string
}

type PPIDCategory struct {
	ID    int64
	Nama  string
	Slug  string
	Items []PPIDItem
}

type PPIDItem struct {
	ID         int64
	KategoriID int64
	Judul      string
	Isi        string
	FileURL    string
	Urutan     int
}

type ContentRepository interface {
	Regulations(ctx context.Context) ([]Regulation, error)
	PublicApps(ctx context.Context) ([]PublicApp, error)
	Videos(ctx context.Context) ([]Video, error)
	PerformanceDocs(ctx context.Context) ([]PerformanceDoc, error)
	ServiceLocations(ctx context.Context) ([]ServiceLocation, error)
	PPID(ctx context.Context) ([]PPIDCategory, error)
}
