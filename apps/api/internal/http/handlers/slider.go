package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"dpmptsp/api/internal/domain"
	"dpmptsp/api/internal/http/middleware"
	"dpmptsp/api/internal/http/render"
)

type Sliders struct {
	Repo domain.SliderRepository
	Log  *slog.Logger
}

type sliderInput struct {
	Gambar    string `json:"gambar"`
	Judul     string `json:"judul"`
	Deskripsi string `json:"deskripsi"`
	Urutan    int    `json:"urutan"`
}

func (h *Sliders) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.Repo.List(r.Context())
	if err != nil {
		render.Error(w, h.Log, middleware.RequestIDFrom(r.Context()), err)
		return
	}
	render.OK(w, toSliderDTOs(rows))
}

func (h *Sliders) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	s, err := h.Repo.ByID(r.Context(), id)
	if err != nil {
		render.Error(w, h.Log, middleware.RequestIDFrom(r.Context()), err)
		return
	}
	render.OK(w, toSliderDTOs([]domain.Slider{*s})[0])
}

func (h *Sliders) Create(w http.ResponseWriter, r *http.Request) {
	in, ok := decodeSlider(w, r)
	if !ok {
		return
	}
	s := &domain.Slider{Gambar: in.Gambar, Judul: in.Judul, Deskripsi: in.Deskripsi, Urutan: in.Urutan}
	if err := h.Repo.Create(r.Context(), s); err != nil {
		render.Error(w, h.Log, middleware.RequestIDFrom(r.Context()), err)
		return
	}
	render.Created(w, toSliderDTOs([]domain.Slider{*s})[0])
}

func (h *Sliders) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	in, ok := decodeSlider(w, r)
	if !ok {
		return
	}
	s := &domain.Slider{ID: id, Gambar: in.Gambar, Judul: in.Judul, Deskripsi: in.Deskripsi, Urutan: in.Urutan}
	if err := h.Repo.Update(r.Context(), s); err != nil {
		render.Error(w, h.Log, middleware.RequestIDFrom(r.Context()), err)
		return
	}
	render.OK(w, toSliderDTOs([]domain.Slider{*s})[0])
}

func (h *Sliders) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	if err := h.Repo.Delete(r.Context(), id); err != nil {
		render.Error(w, h.Log, middleware.RequestIDFrom(r.Context()), err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// decodeSlider reads and validates the body. An image is required: a slide
// without one renders as an empty box across the top of the homepage.
func decodeSlider(w http.ResponseWriter, r *http.Request) (sliderInput, bool) {
	var in sliderInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		render.BadRequest(w, "body must be valid JSON")
		return in, false
	}
	in.Gambar = strings.TrimSpace(in.Gambar)
	if in.Gambar == "" {
		render.JSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "gambar wajib diisi"})
		return in, false
	}
	return in, true
}
