package database

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"dpmptsp/api/internal/domain"
)

type ContentRepo struct{ db *gorm.DB }

func NewContentRepo(db *gorm.DB) *ContentRepo { return &ContentRepo{db: db} }

var _ domain.ContentRepository = (*ContentRepo)(nil)

// listActive is the shape every one of these queries shares: the active rows of
// one table, in display order. Writing it once keeps six repositories from
// being six copies of the same four lines with different table names.
func listActive[M any](ctx context.Context, db *gorm.DB, order string, activeColumn string) ([]M, error) {
	var rows []M
	q := db.WithContext(ctx)
	if activeColumn != "" {
		q = q.Where(activeColumn+" = ?", true)
	}
	if err := q.Order(order).Find(&rows).Error; err != nil {
		var zero M
		return nil, fmt.Errorf("list %T: %w", zero, err)
	}
	return rows, nil
}

// --- models -------------------------------------------------------------

type regulationModel struct {
	ID       int64  `gorm:"column:id;primaryKey"`
	Jenis    string `gorm:"column:jenis"`
	Nomor    string `gorm:"column:nomor"`
	Tahun    int    `gorm:"column:tahun"`
	Judul    string `gorm:"column:judul"`
	File     string `gorm:"column:file"`
	IsActive bool   `gorm:"column:is_active"`
	Urutan   int    `gorm:"column:urutan"`
}

func (regulationModel) TableName() string { return "regulasi" }

type publicAppModel struct {
	ID        int64  `gorm:"column:id;primaryKey"`
	Nama      string `gorm:"column:nama"`
	Deskripsi string `gorm:"column:deskripsi"`
	URL       string `gorm:"column:url"`
	Icon      string `gorm:"column:icon"`
	Target    string `gorm:"column:target"`
	IsActive  bool   `gorm:"column:is_active"`
	Urutan    int    `gorm:"column:urutan"`
}

func (publicAppModel) TableName() string { return "aplikasi_publik" }

type videoModel struct {
	ID        int64  `gorm:"column:id;primaryKey"`
	Judul     string `gorm:"column:judul"`
	Deskripsi string `gorm:"column:deskripsi"`
	Thumbnail string `gorm:"column:thumbnail"`
	URLVideo  string `gorm:"column:url_video"`
	Status    string `gorm:"column:status"`
	Tanggal   string `gorm:"column:tanggal"`
}

func (videoModel) TableName() string { return "videos" }

type performanceDocModel struct {
	ID       int64  `gorm:"column:id;primaryKey"`
	TWOption string `gorm:"column:tw_option"`
	FilePath string `gorm:"column:file_path"`
}

func (performanceDocModel) TableName() string { return "twdata" }

type serviceLocationModel struct {
	ID        int64  `gorm:"column:id;primaryKey"`
	Judul     string `gorm:"column:judul"`
	Deskripsi string `gorm:"column:deskripsi"`
	Gambar    string `gorm:"column:gambar"`
	AltText   string `gorm:"column:alt_text"`
	Lokasi    string `gorm:"column:lokasi"`
	Alamat    string `gorm:"column:alamat"`
	Warna     string `gorm:"column:warna"`
}

func (serviceLocationModel) TableName() string { return "tempat_layanan" }

type ppidCategoryModel struct {
	ID   int64  `gorm:"column:id;primaryKey"`
	Nama string `gorm:"column:nama_kategori"`
	Slug string `gorm:"column:slug"`
}

func (ppidCategoryModel) TableName() string { return "ppid_kategori" }

type ppidItemModel struct {
	ID         int64  `gorm:"column:id;primaryKey"`
	KategoriID int64  `gorm:"column:kategori_id"`
	Judul      string `gorm:"column:judul"`
	Isi        string `gorm:"column:isi"`
	FileURL    string `gorm:"column:file_url"`
	Urutan     int    `gorm:"column:urutan"`
}

func (ppidItemModel) TableName() string { return "ppid_item" }

// --- queries ------------------------------------------------------------

func (r *ContentRepo) Regulations(ctx context.Context) ([]domain.Regulation, error) {
	rows, err := listActive[regulationModel](ctx, r.db, "urutan ASC, tahun DESC, id DESC", "is_active")
	if err != nil {
		return nil, err
	}
	out := make([]domain.Regulation, 0, len(rows))
	for _, m := range rows {
		out = append(out, domain.Regulation{
			ID: m.ID, Jenis: m.Jenis, Nomor: m.Nomor, Tahun: m.Tahun,
			Judul: m.Judul, File: m.File, Urutan: m.Urutan,
		})
	}
	return out, nil
}

func (r *ContentRepo) PublicApps(ctx context.Context) ([]domain.PublicApp, error) {
	rows, err := listActive[publicAppModel](ctx, r.db, "urutan ASC, id ASC", "is_active")
	if err != nil {
		return nil, err
	}
	out := make([]domain.PublicApp, 0, len(rows))
	for _, m := range rows {
		out = append(out, domain.PublicApp{
			ID: m.ID, Nama: m.Nama, Deskripsi: m.Deskripsi, URL: m.URL,
			Icon: m.Icon, Target: m.Target, Urutan: m.Urutan,
		})
	}
	return out, nil
}

func (r *ContentRepo) Videos(ctx context.Context) ([]domain.Video, error) {
	rows, err := listActive[videoModel](ctx, r.db, "tanggal DESC, id DESC", "")
	if err != nil {
		return nil, err
	}
	out := make([]domain.Video, 0, len(rows))
	for _, m := range rows {
		out = append(out, domain.Video{
			ID: m.ID, Judul: m.Judul, Deskripsi: m.Deskripsi,
			Thumbnail: m.Thumbnail, URLVideo: m.URLVideo,
			Status: m.Status, Tanggal: m.Tanggal,
		})
	}
	return out, nil
}

func (r *ContentRepo) PerformanceDocs(ctx context.Context) ([]domain.PerformanceDoc, error) {
	rows, err := listActive[performanceDocModel](ctx, r.db, "id DESC", "")
	if err != nil {
		return nil, err
	}
	out := make([]domain.PerformanceDoc, 0, len(rows))
	for _, m := range rows {
		out = append(out, domain.PerformanceDoc{ID: m.ID, TWOption: m.TWOption, FilePath: m.FilePath})
	}
	return out, nil
}

func (r *ContentRepo) ServiceLocations(ctx context.Context) ([]domain.ServiceLocation, error) {
	rows, err := listActive[serviceLocationModel](ctx, r.db, "id DESC", "")
	if err != nil {
		return nil, err
	}
	out := make([]domain.ServiceLocation, 0, len(rows))
	for _, m := range rows {
		out = append(out, domain.ServiceLocation{
			ID: m.ID, Judul: m.Judul, Deskripsi: m.Deskripsi, Gambar: m.Gambar,
			AltText: m.AltText, Lokasi: m.Lokasi, Alamat: m.Alamat, Warna: m.Warna,
		})
	}
	return out, nil
}

// PPID returns categories with their items already nested, so the page does not
// have to join two lists in the template.
func (r *ContentRepo) PPID(ctx context.Context) ([]domain.PPIDCategory, error) {
	cats, err := listActive[ppidCategoryModel](ctx, r.db, "id ASC", "")
	if err != nil {
		return nil, err
	}
	items, err := listActive[ppidItemModel](ctx, r.db, "kategori_id ASC, urutan ASC, id ASC", "")
	if err != nil {
		return nil, err
	}

	byCategory := make(map[int64][]domain.PPIDItem, len(cats))
	for _, m := range items {
		byCategory[m.KategoriID] = append(byCategory[m.KategoriID], domain.PPIDItem{
			ID: m.ID, KategoriID: m.KategoriID, Judul: m.Judul,
			Isi: m.Isi, FileURL: m.FileURL, Urutan: m.Urutan,
		})
	}

	out := make([]domain.PPIDCategory, 0, len(cats))
	for _, c := range cats {
		out = append(out, domain.PPIDCategory{
			ID: c.ID, Nama: c.Nama, Slug: c.Slug, Items: byCategory[c.ID],
		})
	}
	return out, nil
}
