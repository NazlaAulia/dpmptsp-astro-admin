package handlers

import (
	"log/slog"
	"net/http"

	"dpmptsp/api/internal/domain"
	"dpmptsp/api/internal/http/middleware"
	"dpmptsp/api/internal/http/render"
)

// Content serves the site's ordered content lists.
type Content struct {
	Repo domain.ContentRepository
	Log  *slog.Logger
}

// serve removes the boilerplate that would otherwise be repeated six times:
// call the repository, map the error, wrap the result. Each handler is then
// just its query.
func serve[T any](h *Content, w http.ResponseWriter, r *http.Request, fn func() ([]T, error)) {
	items, err := fn()
	if err != nil {
		render.Error(w, h.Log, middleware.RequestIDFrom(r.Context()), err)
		return
	}
	// Content changes rarely and is read constantly.
	w.Header().Set("Cache-Control", "public, max-age=60")
	render.OK(w, items)
}

// DTOs, for the same reason articles have them: returning a domain struct
// directly puts Go field names on the wire (ID, Judul) and ties the API's shape
// to internal naming.

type regulationDTO struct {
	ID    int64  `json:"id"`
	Jenis string `json:"jenis"`
	Nomor string `json:"nomor"`
	Tahun int    `json:"tahun"`
	Judul string `json:"judul"`
	File  string `json:"file"`
}

type publicAppDTO struct {
	ID        int64  `json:"id"`
	Nama      string `json:"nama"`
	Deskripsi string `json:"deskripsi"`
	URL       string `json:"url"`
	Icon      string `json:"icon"`
	Target    string `json:"target"`
}

type videoDTO struct {
	ID        int64  `json:"id"`
	Judul     string `json:"judul"`
	Deskripsi string `json:"deskripsi"`
	Thumbnail string `json:"thumbnail"`
	URLVideo  string `json:"url_video"`
	Status    string `json:"status"`
	Tanggal   string `json:"tanggal"`
}

type performanceDocDTO struct {
	ID       int64  `json:"id"`
	TWOption string `json:"tw_option"`
	FilePath string `json:"file_path"`
}

type serviceLocationDTO struct {
	ID        int64  `json:"id"`
	Judul     string `json:"judul"`
	Deskripsi string `json:"deskripsi"`
	Gambar    string `json:"gambar"`
	AltText   string `json:"alt_text"`
	Lokasi    string `json:"lokasi"`
	Alamat    string `json:"alamat"`
	Warna     string `json:"warna"`
}

type ppidItemDTO struct {
	ID      int64  `json:"id"`
	Judul   string `json:"judul"`
	Isi     string `json:"isi"`
	FileURL string `json:"file_url"`
}

type ppidCategoryDTO struct {
	ID    int64         `json:"id"`
	Nama  string        `json:"nama_kategori"`
	Slug  string        `json:"slug"`
	Items []ppidItemDTO `json:"items"`
}

func mapSlice[A any, B any](in []A, fn func(A) B) []B {
	out := make([]B, 0, len(in))
	for _, v := range in {
		out = append(out, fn(v))
	}
	return out
}

func (h *Content) Regulations(w http.ResponseWriter, r *http.Request) {
	serve(h, w, r, func() ([]regulationDTO, error) {
		rows, err := h.Repo.Regulations(r.Context())
		return mapSlice(rows, func(m domain.Regulation) regulationDTO {
			return regulationDTO{ID: m.ID, Jenis: m.Jenis, Nomor: m.Nomor, Tahun: m.Tahun, Judul: m.Judul, File: m.File}
		}), err
	})
}

func (h *Content) PublicApps(w http.ResponseWriter, r *http.Request) {
	serve(h, w, r, func() ([]publicAppDTO, error) {
		rows, err := h.Repo.PublicApps(r.Context())
		return mapSlice(rows, func(m domain.PublicApp) publicAppDTO {
			return publicAppDTO{ID: m.ID, Nama: m.Nama, Deskripsi: m.Deskripsi, URL: m.URL, Icon: m.Icon, Target: m.Target}
		}), err
	})
}

func (h *Content) Videos(w http.ResponseWriter, r *http.Request) {
	serve(h, w, r, func() ([]videoDTO, error) {
		rows, err := h.Repo.Videos(r.Context())
		return mapSlice(rows, func(m domain.Video) videoDTO {
			return videoDTO{ID: m.ID, Judul: m.Judul, Deskripsi: m.Deskripsi, Thumbnail: m.Thumbnail, URLVideo: m.URLVideo, Status: m.Status, Tanggal: m.Tanggal}
		}), err
	})
}

func (h *Content) PerformanceDocs(w http.ResponseWriter, r *http.Request) {
	serve(h, w, r, func() ([]performanceDocDTO, error) {
		rows, err := h.Repo.PerformanceDocs(r.Context())
		return mapSlice(rows, func(m domain.PerformanceDoc) performanceDocDTO {
			return performanceDocDTO{ID: m.ID, TWOption: m.TWOption, FilePath: m.FilePath}
		}), err
	})
}

func (h *Content) ServiceLocations(w http.ResponseWriter, r *http.Request) {
	serve(h, w, r, func() ([]serviceLocationDTO, error) {
		rows, err := h.Repo.ServiceLocations(r.Context())
		return mapSlice(rows, func(m domain.ServiceLocation) serviceLocationDTO {
			return serviceLocationDTO{ID: m.ID, Judul: m.Judul, Deskripsi: m.Deskripsi, Gambar: m.Gambar, AltText: m.AltText, Lokasi: m.Lokasi, Alamat: m.Alamat, Warna: m.Warna}
		}), err
	})
}

func (h *Content) PPID(w http.ResponseWriter, r *http.Request) {
	serve(h, w, r, func() ([]ppidCategoryDTO, error) {
		rows, err := h.Repo.PPID(r.Context())
		return mapSlice(rows, func(c domain.PPIDCategory) ppidCategoryDTO {
			return ppidCategoryDTO{
				ID: c.ID, Nama: c.Nama, Slug: c.Slug,
				Items: mapSlice(c.Items, func(i domain.PPIDItem) ppidItemDTO {
					return ppidItemDTO{ID: i.ID, Judul: i.Judul, Isi: i.Isi, FileURL: i.FileURL}
				}),
			}
		}), err
	})
}
