package handlers

import (
	"log/slog"
	"net/http"

	"dpmptsp/api/internal/domain"
	"dpmptsp/api/internal/http/middleware"
	"dpmptsp/api/internal/http/render"
)

// Site serves the branding and navigation every page needs.
type Site struct {
	Repo domain.SiteRepository
	Log  *slog.Logger
}

type menuDTO struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	URL           string    `json:"url"`
	Type          string    `json:"type,omitempty"`
	ContactButton bool      `json:"contact_button"`
	Children      []menuDTO `json:"children,omitempty"`
}

type chromeDTO struct {
	Settings struct {
		Name string `json:"name"`
		Logo string `json:"logo"`
	} `json:"settings"`
	Navigation []menuDTO `json:"navigation"`
	Contact    []menuDTO `json:"contact"`
}

func toMenuDTO(nodes []domain.MenuNode) []menuDTO {
	out := make([]menuDTO, 0, len(nodes))
	for _, n := range nodes {
		out = append(out, menuDTO{
			ID: n.ID, Name: n.Name, URL: n.URL, Type: n.Type,
			ContactButton: n.ContactButton, Children: toMenuDTO(n.Children),
		})
	}
	return out
}

// Chrome handles GET /v1/site/chrome.
func (h *Site) Chrome(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	settings, err := h.Repo.Settings(ctx)
	if err != nil {
		render.Error(w, h.Log, middleware.RequestIDFrom(ctx), err)
		return
	}
	nav, contact, err := h.Repo.MenuTree(ctx)
	if err != nil {
		render.Error(w, h.Log, middleware.RequestIDFrom(ctx), err)
		return
	}

	var out chromeDTO
	out.Settings.Name = settings.Name
	out.Settings.Logo = settings.Logo
	out.Navigation = toMenuDTO(nav)
	out.Contact = toMenuDTO(contact)

	// Site chrome changes rarely and is read constantly.
	w.Header().Set("Cache-Control", "public, max-age=60")
	render.OK(w, out)
}
