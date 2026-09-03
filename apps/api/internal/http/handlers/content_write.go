package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"dpmptsp/api/internal/domain"
	"dpmptsp/api/internal/http/middleware"
	"dpmptsp/api/internal/http/render"
)

// pathID reads a numeric {id} path value.
func pathID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		render.BadRequest(w, "id must be a number")
		return 0, false
	}
	return id, true
}

// decode reads a JSON body into v.
func decode[T any](w http.ResponseWriter, r *http.Request) (T, bool) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		render.BadRequest(w, "body must be valid JSON")
		var zero T
		return zero, false
	}
	return v, true
}

func (h *Content) fail(w http.ResponseWriter, r *http.Request, err error) {
	render.Error(w, h.Log, middleware.RequestIDFrom(r.Context()), err)
}

// --- innovations ---------------------------------------------------------

type innovationDTO struct {
	ID            int64  `json:"id"`
	Slug          string `json:"slug"`
	Nama          string `json:"nama"`
	Singkatan     string `json:"singkatan"`
	Kategori      string `json:"kategori"`
	Deskripsi     string `json:"deskripsi"`
	RancangBangun string `json:"rancang_bangun"`
	Tujuan        string `json:"tujuan"`
	Manfaat       string `json:"manfaat"`
	Hasil         string `json:"hasil"`
	TahunUsulan   int    `json:"tahun_usulan"`
	Tahapan       string `json:"tahapan"`
	Jenis         string `json:"jenis"`
	Gambar        string `json:"gambar"`
	URLLayanan    string `json:"url_layanan"`
	URLLabel      string `json:"url_label"`
	Icon          string `json:"icon"`
	Warna         string `json:"warna"`
	Urutan        int    `json:"urutan"`
	IsActive      bool   `json:"is_active"`
}

func innovationToDTO(v domain.Innovation) innovationDTO {
	return innovationDTO(v)
}

func (d innovationDTO) toDomain() *domain.Innovation {
	v := domain.Innovation(d)
	return &v
}

func (h *Content) Innovations(w http.ResponseWriter, r *http.Request) {
	serve(h, w, r, func() ([]innovationDTO, error) {
		rows, err := h.Repo.Innovations(r.Context())
		return mapSlice(rows, innovationToDTO), err
	})
}

func (h *Content) Innovation(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	v, err := h.Repo.Innovation(r.Context(), id)
	if err != nil {
		h.fail(w, r, err)
		return
	}
	render.OK(w, innovationToDTO(*v))
}

func (h *Content) CreateInnovation(w http.ResponseWriter, r *http.Request) {
	in, ok := decode[innovationDTO](w, r)
	if !ok {
		return
	}
	v := in.toDomain()
	if err := h.Repo.CreateInnovation(r.Context(), v); err != nil {
		h.fail(w, r, err)
		return
	}
	render.Created(w, innovationToDTO(*v))
}

func (h *Content) UpdateInnovation(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	in, ok := decode[innovationDTO](w, r)
	if !ok {
		return
	}
	v := in.toDomain()
	v.ID = id
	if err := h.Repo.UpdateInnovation(r.Context(), v); err != nil {
		h.fail(w, r, err)
		return
	}
	render.OK(w, innovationToDTO(*v))
}

func (h *Content) DeleteInnovation(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	if err := h.Repo.DeleteInnovation(r.Context(), id); err != nil {
		h.fail(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- performance docs ----------------------------------------------------

func (h *Content) PerformanceDoc(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	v, err := h.Repo.PerformanceDoc(r.Context(), id)
	if err != nil {
		h.fail(w, r, err)
		return
	}
	render.OK(w, performanceDocDTO{ID: v.ID, TWOption: v.TWOption, FilePath: v.FilePath})
}

func (h *Content) CreatePerformanceDoc(w http.ResponseWriter, r *http.Request) {
	in, ok := decode[performanceDocDTO](w, r)
	if !ok {
		return
	}
	v := &domain.PerformanceDoc{TWOption: in.TWOption, FilePath: in.FilePath}
	if err := h.Repo.CreatePerformanceDoc(r.Context(), v); err != nil {
		h.fail(w, r, err)
		return
	}
	render.Created(w, performanceDocDTO{ID: v.ID, TWOption: v.TWOption, FilePath: v.FilePath})
}

func (h *Content) UpdatePerformanceDoc(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	in, ok := decode[performanceDocDTO](w, r)
	if !ok {
		return
	}
	v := &domain.PerformanceDoc{ID: id, TWOption: in.TWOption, FilePath: in.FilePath}
	if err := h.Repo.UpdatePerformanceDoc(r.Context(), v); err != nil {
		h.fail(w, r, err)
		return
	}
	render.OK(w, performanceDocDTO{ID: v.ID, TWOption: v.TWOption, FilePath: v.FilePath})
}

func (h *Content) DeletePerformanceDoc(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	if err := h.Repo.DeletePerformanceDoc(r.Context(), id); err != nil {
		h.fail(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- service locations ---------------------------------------------------

func (h *Content) ServiceLocation(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	v, err := h.Repo.ServiceLocation(r.Context(), id)
	if err != nil {
		h.fail(w, r, err)
		return
	}
	render.OK(w, serviceLocationDTO(*v))
}

func (h *Content) CreateServiceLocation(w http.ResponseWriter, r *http.Request) {
	in, ok := decode[serviceLocationDTO](w, r)
	if !ok {
		return
	}
	v := domain.ServiceLocation(in)
	if err := h.Repo.CreateServiceLocation(r.Context(), &v); err != nil {
		h.fail(w, r, err)
		return
	}
	render.Created(w, serviceLocationDTO(v))
}

func (h *Content) UpdateServiceLocation(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	in, ok := decode[serviceLocationDTO](w, r)
	if !ok {
		return
	}
	v := domain.ServiceLocation(in)
	v.ID = id
	if err := h.Repo.UpdateServiceLocation(r.Context(), &v); err != nil {
		h.fail(w, r, err)
		return
	}
	render.OK(w, serviceLocationDTO(v))
}

func (h *Content) DeleteServiceLocation(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	if err := h.Repo.DeleteServiceLocation(r.Context(), id); err != nil {
		h.fail(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- about contents ------------------------------------------------------

// aboutContentDTO is declared in page.go.

func (h *Content) AboutContents(w http.ResponseWriter, r *http.Request) {
	serve(h, w, r, func() ([]aboutContentDTO, error) {
		rows, err := h.Repo.AboutContents(r.Context())
		return mapSlice(rows, func(v domain.AboutContent) aboutContentDTO {
			return aboutContentDTO(v)
		}), err
	})
}

func (h *Content) AboutContent(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	v, err := h.Repo.AboutContentByID(r.Context(), id)
	if err != nil {
		h.fail(w, r, err)
		return
	}
	render.OK(w, aboutContentDTO(*v))
}

func (h *Content) CreateAboutContent(w http.ResponseWriter, r *http.Request) {
	in, ok := decode[aboutContentDTO](w, r)
	if !ok {
		return
	}
	v := domain.AboutContent(in)
	if err := h.Repo.CreateAboutContent(r.Context(), &v); err != nil {
		h.fail(w, r, err)
		return
	}
	render.Created(w, aboutContentDTO(v))
}

func (h *Content) UpdateAboutContent(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	in, ok := decode[aboutContentDTO](w, r)
	if !ok {
		return
	}
	v := domain.AboutContent(in)
	v.ID = id
	if err := h.Repo.UpdateAboutContent(r.Context(), &v); err != nil {
		h.fail(w, r, err)
		return
	}
	render.OK(w, aboutContentDTO(v))
}

func (h *Content) DeleteAboutContent(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	if err := h.Repo.DeleteAboutContent(r.Context(), id); err != nil {
		h.fail(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
