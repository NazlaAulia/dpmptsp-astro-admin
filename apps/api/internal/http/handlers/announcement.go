package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"dpmptsp/api/internal/domain"
	"dpmptsp/api/internal/http/middleware"
	"dpmptsp/api/internal/http/render"
)

type Announcements struct {
	Repo domain.AnnouncementRepository
	Log  *slog.Logger
}

type announcementDTO struct {
	ID        int64  `json:"id"`
	Judul     string `json:"judul"`
	Deskripsi string `json:"deskripsi"`
	Foto      string `json:"foto"`
	LinkURL   string `json:"link_url"`
	IsActive  bool   `json:"is_active"`
	Urutan    int    `json:"urutan"`
	Tipe      string `json:"tipe"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func toAnnouncementDTO(a domain.Announcement) announcementDTO {
	return announcementDTO{
		ID: a.ID, Judul: a.Judul, Deskripsi: a.Deskripsi, Foto: a.Foto,
		LinkURL: a.LinkURL, IsActive: a.IsActive, Urutan: a.Urutan, Tipe: a.Tipe,
		CreatedAt: a.CreatedAt, UpdatedAt: a.UpdatedAt,
	}
}

// List handles GET /v1/announcements. Query parameters: tipe, include_inactive.
func (h *Announcements) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	rows, err := h.Repo.List(r.Context(), domain.AnnouncementFilter{
		Tipe:       q.Get("tipe"),
		ActiveOnly: q.Get("include_inactive") != "true",
	})
	if err != nil {
		render.Error(w, h.Log, middleware.RequestIDFrom(r.Context()), err)
		return
	}
	render.OK(w, mapSlice(rows, toAnnouncementDTO))
}

func (h *Announcements) ByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		render.BadRequest(w, "id must be a number")
		return
	}
	a, err := h.Repo.ByID(r.Context(), id)
	if err != nil {
		render.Error(w, h.Log, middleware.RequestIDFrom(r.Context()), err)
		return
	}
	render.OK(w, toAnnouncementDTO(*a))
}

type announcementInput struct {
	Judul     string `json:"judul"`
	Deskripsi string `json:"deskripsi"`
	Foto      string `json:"foto"`
	LinkURL   string `json:"link_url"`
	IsActive  *bool  `json:"is_active"`
	Urutan    int    `json:"urutan"`
	Tipe      string `json:"tipe"`
}

// toDomain applies defaults and rejects an unknown display type.
func (in announcementInput) toDomain() (*domain.Announcement, error) {
	if in.Judul == "" {
		return nil, errInvalid("judul is required")
	}
	tipe := in.Tipe
	if tipe == "" {
		tipe = domain.AnnouncementNotif
	}
	if tipe != domain.AnnouncementNotif && tipe != domain.AnnouncementModal {
		return nil, errInvalid("tipe must be notif or modal")
	}
	active := true
	if in.IsActive != nil {
		active = *in.IsActive
	}
	return &domain.Announcement{
		Judul: in.Judul, Deskripsi: in.Deskripsi, Foto: in.Foto,
		LinkURL: in.LinkURL, IsActive: active, Urutan: in.Urutan, Tipe: tipe,
	}, nil
}

func (h *Announcements) Create(w http.ResponseWriter, r *http.Request) {
	var in announcementInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		render.BadRequest(w, "body must be valid JSON")
		return
	}
	a, err := in.toDomain()
	if err != nil {
		render.Error(w, h.Log, middleware.RequestIDFrom(r.Context()), err)
		return
	}
	if err := h.Repo.Create(r.Context(), a); err != nil {
		render.Error(w, h.Log, middleware.RequestIDFrom(r.Context()), err)
		return
	}
	render.Created(w, toAnnouncementDTO(*a))
}

func (h *Announcements) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		render.BadRequest(w, "id must be a number")
		return
	}
	var in announcementInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		render.BadRequest(w, "body must be valid JSON")
		return
	}
	a, err := in.toDomain()
	if err != nil {
		render.Error(w, h.Log, middleware.RequestIDFrom(r.Context()), err)
		return
	}
	a.ID = id
	if err := h.Repo.Update(r.Context(), a); err != nil {
		render.Error(w, h.Log, middleware.RequestIDFrom(r.Context()), err)
		return
	}
	render.OK(w, toAnnouncementDTO(*a))
}

func (h *Announcements) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		render.BadRequest(w, "id must be a number")
		return
	}
	if err := h.Repo.Delete(r.Context(), id); err != nil {
		render.Error(w, h.Log, middleware.RequestIDFrom(r.Context()), err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
