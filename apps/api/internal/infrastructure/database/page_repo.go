package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"

	"dpmptsp/api/internal/domain"
)

type PageRepo struct {
	db      *gorm.DB
	content *ContentRepo
	article *ArticleRepo
}

func NewPageRepo(db *gorm.DB, content *ContentRepo, article *ArticleRepo) *PageRepo {
	return &PageRepo{db: db, content: content, article: article}
}

var _ domain.PageRepository = (*PageRepo)(nil)

// --- models -------------------------------------------------------------

type contentBlockModel struct {
	ID     int64  `gorm:"column:id;primaryKey"`
	Tipe   string `gorm:"column:tipe"`
	Konten string `gorm:"column:konten"`
	Urutan int    `gorm:"column:urutan"`
}

type berandaBlockModel struct{ contentBlockModel }

func (berandaBlockModel) TableName() string { return "beranda_dpmptsp" }

type tentangBlockModel struct{ contentBlockModel }

func (tentangBlockModel) TableName() string { return "tentang_dpmptsp" }

type renstraBlockModel struct{ contentBlockModel }

func (renstraBlockModel) TableName() string { return "renstra_investasi" }

type sliderModel struct {
	ID        int64  `gorm:"column:id;primaryKey"`
	Gambar    string `gorm:"column:gambar"`
	Judul     string `gorm:"column:judul"`
	Deskripsi string `gorm:"column:deskripsi"`
	Urutan    int    `gorm:"column:urutan"`
}

type berandaSliderModel struct{ sliderModel }

func (berandaSliderModel) TableName() string { return "beranda_slider" }

type tentangSliderModel struct{ sliderModel }

func (tentangSliderModel) TableName() string { return "tentang_slider" }

type advantageModel struct {
	ID        int64     `gorm:"column:id;primaryKey"`
	Judul     string    `gorm:"column:judul"`
	Deskripsi string    `gorm:"column:deskripsi"`
	Icon      string    `gorm:"column:icon"`
	WarnaIcon string    `gorm:"column:warna_icon"`
	WarnaBg   string    `gorm:"column:warna_bg"`
	Link      string    `gorm:"column:link"`
	Tanggal   time.Time `gorm:"column:tanggal"`
}

func (advantageModel) TableName() string { return "keunggulan" }

type notificationModel struct {
	ID       int64      `gorm:"column:id;primaryKey"`
	Judul    string     `gorm:"column:judul"`
	Isi      string     `gorm:"column:isi"`
	Gambar   string     `gorm:"column:gambar"`
	Tipe     string     `gorm:"column:tipe"`
	Icon     string     `gorm:"column:icon"`
	Link     string     `gorm:"column:link"`
	TeksLink string     `gorm:"column:teks_link"`
	IsActive bool       `gorm:"column:is_active"`
	Mulai    *time.Time `gorm:"column:mulai"`
	Selesai  *time.Time `gorm:"column:selesai"`
	Urutan   int        `gorm:"column:urutan"`
}

func (notificationModel) TableName() string { return "notifikasi_beranda" }

type investmentMapModel struct {
	ID               int64  `gorm:"column:id;primaryKey"`
	Judul            string `gorm:"column:judul"`
	Deskripsi        string `gorm:"column:deskripsi"`
	Gambar           string `gorm:"column:gambar"`
	Link             string `gorm:"column:link"`
	JudulRenstra     string `gorm:"column:judul_renstra"`
	DeskripsiRenstra string `gorm:"column:deskripsi_renstra"`
}

func (investmentMapModel) TableName() string { return "peta_investasi" }

type aboutContentModel struct {
	ID         int64  `gorm:"column:id;primaryKey"`
	Nama       string `gorm:"column:nama"`
	Isi        string `gorm:"column:isi"`
	Foto       string `gorm:"column:foto"`
	Judul      string `gorm:"column:judul"`
	Keterangan string `gorm:"column:keterangan"`
	Tags       string `gorm:"column:tags"`
}

func (aboutContentModel) TableName() string { return "tentang_kami" }

type serviceStandardModel struct {
	ID        int64  `gorm:"column:id;primaryKey"`
	TabKey    string `gorm:"column:tab_key"`
	TabIcon   string `gorm:"column:tab_icon"`
	TabJudul  string `gorm:"column:tab_judul"`
	Judul     string `gorm:"column:judul"`
	Deskripsi string `gorm:"column:deskripsi"`
	Isi       string `gorm:"column:isi"`
	Urutan    int    `gorm:"column:urutan"`
	IsActive  bool   `gorm:"column:is_active"`
}

func (serviceStandardModel) TableName() string { return "pelayanan_standar" }

type infoSectionModel struct {
	ID                 int64  `gorm:"column:id;primaryKey"`
	SectionID          string `gorm:"column:section_id"`
	Badge              string `gorm:"column:badge"`
	Judul              string `gorm:"column:judul"`
	JudulHighlight     string `gorm:"column:judul_highlight"`
	Deskripsi          string `gorm:"column:deskripsi"`
	Icon               string `gorm:"column:icon"`
	Subjudul           string `gorm:"column:subjudul"`
	SubjudulKeterangan string `gorm:"column:subjudul_keterangan"`
	Isi                string `gorm:"column:isi"`
	AksesJudul         string `gorm:"column:akses_judul"`
	AksesDeskripsi     string `gorm:"column:akses_deskripsi"`
	AksesButton        string `gorm:"column:akses_button"`
	AksesLink          string `gorm:"column:akses_link"`
	EmailLabel         string `gorm:"column:email_label"`
	Email              string `gorm:"column:email"`
	EmailButton        string `gorm:"column:email_button"`
	AlamatLabel        string `gorm:"column:alamat_label"`
	NamaInstansi       string `gorm:"column:nama_instansi"`
	Alamat             string `gorm:"column:alamat"`
	CTAJudul           string `gorm:"column:cta_judul"`
	CTADeskripsi       string `gorm:"column:cta_deskripsi"`
	CTAButton          string `gorm:"column:cta_button"`
	CTALink            string `gorm:"column:cta_link"`
	IsActive           bool   `gorm:"column:is_active"`
}

func (infoSectionModel) TableName() string { return "informasi_section" }

type infoSectionItemModel struct {
	ID        int64  `gorm:"column:id;primaryKey"`
	SectionID string `gorm:"column:section_id"`
	Icon      string `gorm:"column:icon"`
	Judul     string `gorm:"column:judul"`
	Deskripsi string `gorm:"column:deskripsi"`
	Warna     string `gorm:"column:warna"`
	Urutan    int    `gorm:"column:urutan"`
	IsActive  bool   `gorm:"column:is_active"`
}

func (infoSectionItemModel) TableName() string { return "informasi_section_item" }

type photoModel struct {
	ID        int64     `gorm:"column:id;primaryKey"`
	Title     string    `gorm:"column:title"`
	ImagePath string    `gorm:"column:image_path"`
	AltText   string    `gorm:"column:alt_text"`
	Location  string    `gorm:"column:location"`
	Alamat    string    `gorm:"column:alamat"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (photoModel) TableName() string { return "tempat_pelayanan_dpmptsp" }

// --- queries ------------------------------------------------------------

// Home fetches every section concurrently.
func (r *PageRepo) Home(ctx context.Context) (domain.HomeContent, error) {
	var out domain.HomeContent
	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		var rows []notificationModel
		now := time.Now()
		err := r.db.WithContext(gctx).
			Where("is_active = ?", true).
			Where("mulai IS NULL OR mulai <= ?", now).
			Where("selesai IS NULL OR selesai >= ?", now).
			Order("urutan ASC, id DESC").Limit(5).Find(&rows).Error
		if err != nil {
			return fmt.Errorf("notifications: %w", err)
		}
		for _, m := range rows {
			out.Notifications = append(out.Notifications, domain.Notification{
				ID: m.ID, Judul: m.Judul, Isi: m.Isi, Gambar: m.Gambar,
				Tipe: m.Tipe, Icon: m.Icon, Link: m.Link, TeksLink: m.TeksLink,
			})
		}
		return nil
	})

	g.Go(func() error {
		var rows []berandaBlockModel
		if err := r.db.WithContext(gctx).Order("urutan ASC").Find(&rows).Error; err != nil {
			return fmt.Errorf("home blocks: %w", err)
		}
		for _, m := range rows {
			out.Blocks = append(out.Blocks, toBlock(m.contentBlockModel))
		}
		return nil
	})

	g.Go(func() error {
		var rows []berandaSliderModel
		if err := r.db.WithContext(gctx).Order("urutan ASC, id ASC").Find(&rows).Error; err != nil {
			return fmt.Errorf("home sliders: %w", err)
		}
		for _, m := range rows {
			out.Sliders = append(out.Sliders, toSlider(m.sliderModel))
		}
		return nil
	})

	g.Go(func() error {
		var rows []advantageModel
		if err := r.db.WithContext(gctx).Order("tanggal ASC").Limit(6).Find(&rows).Error; err != nil {
			return fmt.Errorf("advantages: %w", err)
		}
		for _, m := range rows {
			out.Advantages = append(out.Advantages, domain.Advantage{
				ID: m.ID, Judul: m.Judul, Deskripsi: m.Deskripsi, Icon: m.Icon,
				WarnaIcon: m.WarnaIcon, WarnaBg: m.WarnaBg, Link: m.Link,
			})
		}
		return nil
	})

	g.Go(func() error {
		var m investmentMapModel
		err := r.db.WithContext(gctx).Order("id ASC").First(&m).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil // optional section
		}
		if err != nil {
			return fmt.Errorf("investment map: %w", err)
		}
		out.InvestmentMap = &domain.InvestmentMap{
			ID: m.ID, Judul: m.Judul, Deskripsi: m.Deskripsi, Gambar: m.Gambar,
			Link: m.Link, JudulRenstra: m.JudulRenstra, DeskripsiRenstra: m.DeskripsiRenstra,
		}
		return nil
	})

	g.Go(func() error {
		var rows []renstraBlockModel
		if err := r.db.WithContext(gctx).Order("urutan ASC, id ASC").Find(&rows).Error; err != nil {
			return fmt.Errorf("renstra: %w", err)
		}
		for _, m := range rows {
			out.Renstra = append(out.Renstra, toBlock(m.contentBlockModel))
		}
		return nil
	})

	g.Go(func() error {
		v, err := r.content.Regulations(gctx)
		out.Regulations = v
		return err
	})
	g.Go(func() error {
		v, err := r.content.Videos(gctx)
		out.Videos = v
		return err
	})
	g.Go(func() error {
		v, err := r.content.PublicApps(gctx)
		out.PublicApps = v
		return err
	})
	g.Go(func() error {
		page, err := r.article.List(gctx, domain.ArticleFilter{ActiveOnly: true, Limit: 6})
		out.Articles = page.Items
		return err
	})

	return out, g.Wait()
}

func (r *PageRepo) About(ctx context.Context) (domain.AboutContentPage, error) {
	var out domain.AboutContentPage
	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		var rows []tentangBlockModel
		if err := r.db.WithContext(gctx).Order("urutan ASC").Find(&rows).Error; err != nil {
			return fmt.Errorf("about blocks: %w", err)
		}
		for _, m := range rows {
			out.Blocks = append(out.Blocks, toBlock(m.contentBlockModel))
		}
		return nil
	})

	g.Go(func() error {
		var rows []tentangSliderModel
		if err := r.db.WithContext(gctx).Order("urutan ASC").Find(&rows).Error; err != nil {
			return fmt.Errorf("about sliders: %w", err)
		}
		for _, m := range rows {
			out.Sliders = append(out.Sliders, toSlider(m.sliderModel))
		}
		return nil
	})

	g.Go(func() error {
		// The legacy page deduplicated by name with a MIN(id) self-join,
		// because the table has repeated entries. Same result, expressed as a
		// subquery rather than a join the ORM would have to be talked into.
		sub := r.db.WithContext(gctx).Model(&aboutContentModel{}).
			Select("MIN(id)").Group("nama")
		var rows []aboutContentModel
		if err := r.db.WithContext(gctx).Where("id IN (?)", sub).Order("id ASC").Find(&rows).Error; err != nil {
			return fmt.Errorf("about contents: %w", err)
		}
		for _, m := range rows {
			out.Contents = append(out.Contents, domain.AboutContent{
				ID: m.ID, Nama: m.Nama, Isi: m.Isi, Foto: m.Foto,
				Judul: m.Judul, Keterangan: m.Keterangan, Tags: m.Tags,
			})
		}
		return nil
	})

	return out, g.Wait()
}

func (r *PageRepo) ServiceStandards(ctx context.Context) ([]domain.ServiceStandard, error) {
	var rows []serviceStandardModel
	if err := r.db.WithContext(ctx).Where("is_active = ?", true).
		Order("urutan ASC, id ASC").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("service standards: %w", err)
	}
	out := make([]domain.ServiceStandard, 0, len(rows))
	for _, m := range rows {
		out = append(out, domain.ServiceStandard{
			ID: m.ID, TabKey: m.TabKey, TabIcon: m.TabIcon, TabJudul: m.TabJudul,
			Judul: m.Judul, Deskripsi: m.Deskripsi, Isi: m.Isi, Urutan: m.Urutan,
		})
	}
	return out, nil
}

func (r *PageRepo) InfoSection(ctx context.Context, sectionID string) (*domain.InfoSection, error) {
	var m infoSectionModel
	err := r.db.WithContext(ctx).Where("section_id = ? AND is_active = ?", sectionID, true).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("info section: %w", err)
	}

	var items []infoSectionItemModel
	if err := r.db.WithContext(ctx).Where("section_id = ? AND is_active = ?", sectionID, true).
		Order("urutan ASC, id ASC").Find(&items).Error; err != nil {
		return nil, fmt.Errorf("info section items: %w", err)
	}

	out := domain.InfoSection{
		ID: m.ID, SectionID: m.SectionID, Badge: m.Badge, Judul: m.Judul,
		JudulHighlight: m.JudulHighlight, Deskripsi: m.Deskripsi, Icon: m.Icon,
		Subjudul: m.Subjudul, SubjudulKeterangan: m.SubjudulKeterangan, Isi: m.Isi,
		AksesJudul: m.AksesJudul, AksesDeskripsi: m.AksesDeskripsi,
		AksesButton: m.AksesButton, AksesLink: m.AksesLink,
		EmailLabel: m.EmailLabel, Email: m.Email, EmailButton: m.EmailButton,
		AlamatLabel: m.AlamatLabel, NamaInstansi: m.NamaInstansi, Alamat: m.Alamat,
		CTAJudul: m.CTAJudul, CTADeskripsi: m.CTADeskripsi,
		CTAButton: m.CTAButton, CTALink: m.CTALink,
	}
	for _, i := range items {
		out.Items = append(out.Items, domain.InfoSectionItem{
			ID: i.ID, Icon: i.Icon, Judul: i.Judul, Deskripsi: i.Deskripsi, Warna: i.Warna,
		})
	}
	return &out, nil
}

func (r *PageRepo) Photos(ctx context.Context, limit, offset int) (domain.Page[domain.Photo], error) {
	var out domain.Page[domain.Photo]
	if err := r.db.WithContext(ctx).Model(&photoModel{}).Count(&out.Total).Error; err != nil {
		return out, fmt.Errorf("count photos: %w", err)
	}
	var rows []photoModel
	if err := r.db.WithContext(ctx).Order("created_at DESC").
		Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return out, fmt.Errorf("list photos: %w", err)
	}
	out.Items = make([]domain.Photo, 0, len(rows))
	for _, m := range rows {
		out.Items = append(out.Items, domain.Photo{
			ID: m.ID, Title: m.Title, ImagePath: m.ImagePath, AltText: m.AltText,
			Location: m.Location, Alamat: m.Alamat,
			CreatedAt: m.CreatedAt.Format(time.RFC3339),
		})
	}
	return out, nil
}

func toBlock(m contentBlockModel) domain.ContentBlock {
	return domain.ContentBlock{ID: m.ID, Tipe: m.Tipe, Konten: m.Konten, Urutan: m.Urutan}
}

func toSlider(m sliderModel) domain.Slider {
	return domain.Slider{ID: m.ID, Gambar: m.Gambar, Judul: m.Judul, Deskripsi: m.Deskripsi, Urutan: m.Urutan}
}
