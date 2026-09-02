package domain

import "context"

// ContentBlock is a piece of ordered page copy: a heading, a paragraph, a link.
// beranda_dpmptsp, tentang_dpmptsp and renstra_investasi all share this shape.
type ContentBlock struct {
	ID     int64
	Tipe   string
	Konten string
	Urutan int
}

// Slider is a carousel entry. beranda_slider and tentang_slider share this.
type Slider struct {
	ID        int64
	Gambar    string
	Judul     string
	Deskripsi string
	Urutan    int
}

type Advantage struct {
	ID        int64
	Judul     string
	Deskripsi string
	Icon      string
	WarnaIcon string
	WarnaBg   string
	Link      string
}

type Notification struct {
	ID       int64
	Judul    string
	Isi      string
	Gambar   string
	Tipe     string
	Icon     string
	Link     string
	TeksLink string
}

type InvestmentMap struct {
	ID               int64
	Judul            string
	Deskripsi        string
	Gambar           string
	Link             string
	JudulRenstra     string
	DeskripsiRenstra string
}

// AboutContent is one entry of the about page's tab content (tentang_kami).
type AboutContent struct {
	ID         int64
	Nama       string
	Isi        string
	Foto       string
	Judul      string
	Keterangan string
	Tags       string
}

type ServiceStandard struct {
	ID        int64
	TabKey    string
	TabIcon   string
	TabJudul  string
	Judul     string
	Deskripsi string
	// Isi holds a JSON document the page parses. Kept as a string here: it is
	// never queried by path, so decoding it in the API would only mean encoding
	// it again on the way out.
	Isi    string
	Urutan int
}

// InfoSection is a single configurable page section plus its items.
type InfoSection struct {
	ID                 int64
	SectionID          string
	Badge              string
	Judul              string
	JudulHighlight     string
	Deskripsi          string
	Icon               string
	Subjudul           string
	SubjudulKeterangan string
	Isi                string
	AksesJudul         string
	AksesDeskripsi     string
	AksesButton        string
	AksesLink          string
	EmailLabel         string
	Email              string
	EmailButton        string
	AlamatLabel        string
	NamaInstansi       string
	Alamat             string
	CTAJudul           string
	CTADeskripsi       string
	CTAButton          string
	CTALink            string
	Items              []InfoSectionItem
}

type InfoSectionItem struct {
	ID        int64
	Icon      string
	Judul     string
	Deskripsi string
	Warna     string
}

type Photo struct {
	ID        int64
	Title     string
	ImagePath string
	AltText   string
	Location  string
	Alamat    string
	CreatedAt string
}

// HomeContent is everything the homepage renders, in one object.
//
// The page it replaces ran ten sequential queries. One composite means one
// round trip and one cache entry, which matters because this is the most
// requested page on the site.
type HomeContent struct {
	Notifications []Notification
	Blocks        []ContentBlock
	Regulations   []Regulation
	Sliders       []Slider
	Articles      []Article
	Advantages    []Advantage
	Videos        []Video
	InvestmentMap *InvestmentMap
	Renstra       []ContentBlock
	PublicApps    []PublicApp
}

// AboutContentPage is the about page's data.
type AboutContentPage struct {
	Blocks   []ContentBlock
	Sliders  []Slider
	Contents []AboutContent
}

type PageRepository interface {
	Home(ctx context.Context) (HomeContent, error)
	About(ctx context.Context) (AboutContentPage, error)
	ServiceStandards(ctx context.Context) ([]ServiceStandard, error)
	InfoSection(ctx context.Context, sectionID string) (*InfoSection, error)
	Photos(ctx context.Context, limit, offset int) (Page[Photo], error)
}
