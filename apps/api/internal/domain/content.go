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

// Innovation is an inovasi_layanan row.
type Innovation struct {
	ID            int64
	Slug          string
	Nama          string
	Singkatan     string
	Kategori      string
	Deskripsi     string
	RancangBangun string
	Tujuan        string
	Manfaat       string
	Hasil         string
	TahunUsulan   int
	Tahapan       string
	Jenis         string
	Gambar        string
	URLLayanan    string
	URLLabel      string
	Icon          string
	Warna         string
	Urutan        int
	IsActive      bool
}

type ContentRepository interface {
	Regulations(ctx context.Context) ([]Regulation, error)
	PublicApps(ctx context.Context) ([]PublicApp, error)
	Videos(ctx context.Context) ([]Video, error)
	PerformanceDocs(ctx context.Context) ([]PerformanceDoc, error)
	ServiceLocations(ctx context.Context) ([]ServiceLocation, error)
	PPID(ctx context.Context) ([]PPIDCategory, error)

	Innovations(ctx context.Context) ([]Innovation, error)
	Innovation(ctx context.Context, id int64) (*Innovation, error)
	CreateInnovation(ctx context.Context, v *Innovation) error
	UpdateInnovation(ctx context.Context, v *Innovation) error
	DeleteInnovation(ctx context.Context, id int64) error

	PerformanceDoc(ctx context.Context, id int64) (*PerformanceDoc, error)
	CreatePerformanceDoc(ctx context.Context, v *PerformanceDoc) error
	UpdatePerformanceDoc(ctx context.Context, v *PerformanceDoc) error
	DeletePerformanceDoc(ctx context.Context, id int64) error

	ServiceLocation(ctx context.Context, id int64) (*ServiceLocation, error)
	CreateServiceLocation(ctx context.Context, v *ServiceLocation) error
	UpdateServiceLocation(ctx context.Context, v *ServiceLocation) error
	DeleteServiceLocation(ctx context.Context, id int64) error

	AboutContents(ctx context.Context) ([]AboutContent, error)
	AboutContentByID(ctx context.Context, id int64) (*AboutContent, error)
	CreateAboutContent(ctx context.Context, v *AboutContent) error
	UpdateAboutContent(ctx context.Context, v *AboutContent) error
	DeleteAboutContent(ctx context.Context, id int64) error
}
