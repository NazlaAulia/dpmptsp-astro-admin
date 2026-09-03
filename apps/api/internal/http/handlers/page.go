package handlers

import (
	"log/slog"
	"net/http"
	"strconv"

	"dpmptsp/api/internal/domain"
	"dpmptsp/api/internal/http/middleware"
	"dpmptsp/api/internal/http/render"
)

// Pages serves the composite page payloads.
type Pages struct {
	Repo domain.PageRepository
	Log  *slog.Logger
}

type blockDTO struct {
	ID     int64  `json:"id"`
	Tipe   string `json:"tipe"`
	Konten string `json:"konten"`
	Urutan int    `json:"urutan"`
}

type sliderDTO struct {
	ID        int64  `json:"id"`
	Gambar    string `json:"gambar"`
	Judul     string `json:"judul"`
	Deskripsi string `json:"deskripsi"`
	Urutan    int    `json:"urutan"`
}

type advantageDTO struct {
	ID        int64  `json:"id"`
	Judul     string `json:"judul"`
	Deskripsi string `json:"deskripsi"`
	Icon      string `json:"icon"`
	WarnaIcon string `json:"warna_icon"`
	WarnaBg   string `json:"warna_bg"`
	Link      string `json:"link"`
}

type notificationDTO struct {
	ID       int64  `json:"id"`
	Judul    string `json:"judul"`
	Isi      string `json:"isi"`
	Gambar   string `json:"gambar"`
	Tipe     string `json:"tipe"`
	Icon     string `json:"icon"`
	Link     string `json:"link"`
	TeksLink string `json:"teks_link"`
}

type investmentMapDTO struct {
	ID               int64  `json:"id"`
	Judul            string `json:"judul"`
	Deskripsi        string `json:"deskripsi"`
	Gambar           string `json:"gambar"`
	Link             string `json:"link"`
	JudulRenstra     string `json:"judul_renstra"`
	DeskripsiRenstra string `json:"deskripsi_renstra"`
}

type homeDTO struct {
	Notifications []notificationDTO `json:"notifications"`
	Blocks        []blockDTO        `json:"blocks"`
	Regulations   []regulationDTO   `json:"regulations"`
	Sliders       []sliderDTO       `json:"sliders"`
	Articles      []articleDTO      `json:"articles"`
	Advantages    []advantageDTO    `json:"advantages"`
	Videos        []videoDTO        `json:"videos"`
	InvestmentMap *investmentMapDTO `json:"investment_map"`
	Renstra       []blockDTO        `json:"renstra"`
	PublicApps    []publicAppDTO    `json:"public_apps"`
}

func toBlockDTOs(in []domain.ContentBlock) []blockDTO {
	return mapSlice(in, func(b domain.ContentBlock) blockDTO {
		return blockDTO{ID: b.ID, Tipe: b.Tipe, Konten: b.Konten, Urutan: b.Urutan}
	})
}

func toSliderDTOs(in []domain.Slider) []sliderDTO {
	return mapSlice(in, func(s domain.Slider) sliderDTO {
		return sliderDTO{ID: s.ID, Gambar: s.Gambar, Judul: s.Judul, Deskripsi: s.Deskripsi, Urutan: s.Urutan}
	})
}

// Home handles GET /v1/home.
func (h *Pages) Home(w http.ResponseWriter, r *http.Request) {
	data, err := h.Repo.Home(r.Context())
	if err != nil {
		render.Error(w, h.Log, middleware.RequestIDFrom(r.Context()), err)
		return
	}

	out := homeDTO{
		Blocks:  toBlockDTOs(data.Blocks),
		Sliders: toSliderDTOs(data.Sliders),
		Renstra: toBlockDTOs(data.Renstra),
		Notifications: mapSlice(data.Notifications, func(n domain.Notification) notificationDTO {
			return notificationDTO{ID: n.ID, Judul: n.Judul, Isi: n.Isi, Gambar: n.Gambar,
				Tipe: n.Tipe, Icon: n.Icon, Link: n.Link, TeksLink: n.TeksLink}
		}),
		Regulations: mapSlice(data.Regulations, func(m domain.Regulation) regulationDTO {
			return regulationDTO{ID: m.ID, Jenis: m.Jenis, Nomor: m.Nomor, Tahun: m.Tahun, Judul: m.Judul, File: m.File}
		}),
		Articles: mapSlice(data.Articles, func(a domain.Article) articleDTO { return toDTO(a, false) }),
		Advantages: mapSlice(data.Advantages, func(a domain.Advantage) advantageDTO {
			return advantageDTO{ID: a.ID, Judul: a.Judul, Deskripsi: a.Deskripsi, Icon: a.Icon,
				WarnaIcon: a.WarnaIcon, WarnaBg: a.WarnaBg, Link: a.Link}
		}),
		Videos: mapSlice(data.Videos, func(m domain.Video) videoDTO {
			return videoDTO{ID: m.ID, Judul: m.Judul, Deskripsi: m.Deskripsi, Thumbnail: m.Thumbnail,
				URLVideo: m.URLVideo, Status: m.Status, Tanggal: m.Tanggal}
		}),
		PublicApps: mapSlice(data.PublicApps, func(m domain.PublicApp) publicAppDTO {
			return publicAppDTO{ID: m.ID, Nama: m.Nama, Deskripsi: m.Deskripsi, URL: m.URL, Icon: m.Icon, Target: m.Target}
		}),
	}
	if data.InvestmentMap != nil {
		m := data.InvestmentMap
		out.InvestmentMap = &investmentMapDTO{ID: m.ID, Judul: m.Judul, Deskripsi: m.Deskripsi,
			Gambar: m.Gambar, Link: m.Link, JudulRenstra: m.JudulRenstra, DeskripsiRenstra: m.DeskripsiRenstra}
	}

	w.Header().Set("Cache-Control", "public, max-age=60")
	render.OK(w, out)
}

type aboutContentDTO struct {
	ID         int64  `json:"id"`
	Nama       string `json:"nama"`
	Isi        string `json:"isi"`
	Foto       string `json:"foto"`
	Judul      string `json:"judul"`
	Keterangan string `json:"keterangan"`
	Tags       string `json:"tags"`
}

type aboutDTO struct {
	Blocks   []blockDTO        `json:"blocks"`
	Sliders  []sliderDTO       `json:"sliders"`
	Contents []aboutContentDTO `json:"contents"`
}

func (h *Pages) About(w http.ResponseWriter, r *http.Request) {
	data, err := h.Repo.About(r.Context())
	if err != nil {
		render.Error(w, h.Log, middleware.RequestIDFrom(r.Context()), err)
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=60")
	render.OK(w, aboutDTO{
		Blocks:  toBlockDTOs(data.Blocks),
		Sliders: toSliderDTOs(data.Sliders),
		Contents: mapSlice(data.Contents, func(c domain.AboutContent) aboutContentDTO {
			return aboutContentDTO{ID: c.ID, Nama: c.Nama, Isi: c.Isi, Foto: c.Foto,
				Judul: c.Judul, Keterangan: c.Keterangan, Tags: c.Tags}
		}),
	})
}

type serviceStandardDTO struct {
	ID        int64  `json:"id"`
	TabKey    string `json:"tab_key"`
	TabIcon   string `json:"tab_icon"`
	TabJudul  string `json:"tab_judul"`
	Judul     string `json:"judul"`
	Deskripsi string `json:"deskripsi"`
	Isi       string `json:"isi"`
	Urutan    int    `json:"urutan"`
}

func (h *Pages) ServiceStandards(w http.ResponseWriter, r *http.Request) {
	rows, err := h.Repo.ServiceStandards(r.Context())
	if err != nil {
		render.Error(w, h.Log, middleware.RequestIDFrom(r.Context()), err)
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=60")
	render.OK(w, mapSlice(rows, func(m domain.ServiceStandard) serviceStandardDTO {
		return serviceStandardDTO{ID: m.ID, TabKey: m.TabKey, TabIcon: m.TabIcon,
			TabJudul: m.TabJudul, Judul: m.Judul, Deskripsi: m.Deskripsi, Isi: m.Isi, Urutan: m.Urutan}
	}))
}

type infoSectionItemDTO struct {
	ID        int64  `json:"id"`
	Icon      string `json:"icon"`
	Judul     string `json:"judul"`
	Deskripsi string `json:"deskripsi"`
	Warna     string `json:"warna"`
}

type infoSectionDTO struct {
	ID                 int64                `json:"id"`
	SectionID          string               `json:"section_id"`
	Badge              string               `json:"badge"`
	Judul              string               `json:"judul"`
	JudulHighlight     string               `json:"judul_highlight"`
	Deskripsi          string               `json:"deskripsi"`
	Icon               string               `json:"icon"`
	Subjudul           string               `json:"subjudul"`
	SubjudulKeterangan string               `json:"subjudul_keterangan"`
	Isi                string               `json:"isi"`
	AksesJudul         string               `json:"akses_judul"`
	AksesDeskripsi     string               `json:"akses_deskripsi"`
	AksesButton        string               `json:"akses_button"`
	AksesLink          string               `json:"akses_link"`
	EmailLabel         string               `json:"email_label"`
	Email              string               `json:"email"`
	EmailButton        string               `json:"email_button"`
	AlamatLabel        string               `json:"alamat_label"`
	NamaInstansi       string               `json:"nama_instansi"`
	Alamat             string               `json:"alamat"`
	CTAJudul           string               `json:"cta_judul"`
	CTADeskripsi       string               `json:"cta_deskripsi"`
	CTAButton          string               `json:"cta_button"`
	CTALink            string               `json:"cta_link"`
	Items              []infoSectionItemDTO `json:"items"`
}

func (h *Pages) InfoSection(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("sectionId")
	s, err := h.Repo.InfoSection(r.Context(), id)
	if err != nil {
		render.Error(w, h.Log, middleware.RequestIDFrom(r.Context()), err)
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=60")
	render.OK(w, infoSectionDTO{
		ID: s.ID, SectionID: s.SectionID, Badge: s.Badge, Judul: s.Judul,
		JudulHighlight: s.JudulHighlight, Deskripsi: s.Deskripsi, Icon: s.Icon,
		Subjudul: s.Subjudul, SubjudulKeterangan: s.SubjudulKeterangan, Isi: s.Isi,
		AksesJudul: s.AksesJudul, AksesDeskripsi: s.AksesDeskripsi,
		AksesButton: s.AksesButton, AksesLink: s.AksesLink,
		EmailLabel: s.EmailLabel, Email: s.Email, EmailButton: s.EmailButton,
		AlamatLabel: s.AlamatLabel, NamaInstansi: s.NamaInstansi, Alamat: s.Alamat,
		CTAJudul: s.CTAJudul, CTADeskripsi: s.CTADeskripsi,
		CTAButton: s.CTAButton, CTALink: s.CTALink,
		Items: mapSlice(s.Items, func(i domain.InfoSectionItem) infoSectionItemDTO {
			return infoSectionItemDTO{ID: i.ID, Icon: i.Icon, Judul: i.Judul, Deskripsi: i.Deskripsi, Warna: i.Warna}
		}),
	})
}

type photoDTO struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	ImagePath string `json:"image_path"`
	AltText   string `json:"alt_text"`
	Location  string `json:"location"`
	Alamat    string `json:"alamat"`
	CreatedAt string `json:"created_at"`
}

func (h *Pages) Photos(w http.ResponseWriter, r *http.Request) {
	page := atoiDefault(r.URL.Query().Get("page"), 1)
	if page < 1 {
		page = 1
	}
	perPage := atoiDefault(r.URL.Query().Get("per_page"), 8)
	if perPage < 1 || perPage > 100 {
		perPage = 8
	}

	res, err := h.Repo.Photos(r.Context(), perPage, (page-1)*perPage)
	if err != nil {
		render.Error(w, h.Log, middleware.RequestIDFrom(r.Context()), err)
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=60")
	render.Paged(w, mapSlice(res.Items, func(p domain.Photo) photoDTO {
		return photoDTO{ID: p.ID, Title: p.Title, ImagePath: p.ImagePath,
			AltText: p.AltText, Location: p.Location, Alamat: p.Alamat, CreatedAt: p.CreatedAt}
	}), res.Total, page, perPage)
}

var _ = strconv.Itoa
