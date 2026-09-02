package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"dpmptsp/api/internal/application"
	"dpmptsp/api/internal/domain"
	"dpmptsp/api/internal/http/middleware"
	"dpmptsp/api/internal/http/render"
)

type Articles struct {
	Service *application.ArticleService
	Log     *slog.Logger
}

// articleDTO is the wire shape. It is deliberately separate from the domain
// type: renaming a domain field must not silently change the API, and internal
// fields must not leak just because someone added them to the struct.
type articleDTO struct {
	ID          int64     `json:"id"`
	CategoryID  int64     `json:"category_id"`
	Category    string    `json:"category,omitempty"`
	Slug        string    `json:"slug"`
	Title       string    `json:"title"`
	Content     string    `json:"content,omitempty"`
	RefContent  string    `json:"ref_content,omitempty"`
	Excerpt     string    `json:"excerpt,omitempty"`
	Picture     string    `json:"picture,omitempty"`
	PublishedAt time.Time `json:"published_at"`
	IsHeadline  bool      `json:"is_headline"`
	Hits        int64     `json:"hits"`
}

func toDTO(a domain.Article, withContent bool) articleDTO {
	d := articleDTO{
		ID: a.ID, CategoryID: a.CategoryID, Slug: a.Slug, Title: a.Title,
		Picture: a.Picture, PublishedAt: a.PublishedAt,
		IsHeadline: a.IsHeadline, Hits: a.Hits,
	}
	if a.Category != nil {
		d.Category = a.Category.Title
	}
	if withContent {
		d.Content = a.Content
		// The article page renders ref_content, not content. Both columns
		// exist in the legacy schema and hold different markup.
		d.RefContent = a.RefContent
	} else {
		// A list of 10 articles must not ship 10 full HTML bodies. The legacy
		// page did exactly that and it is why one request moved 1.7 MB.
		d.Excerpt = excerpt(a.Content, 200)
	}
	return d
}

// List handles GET /v1/articles.
func (h *Articles) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	page := atoiDefault(q.Get("page"), 1)
	if page < 1 {
		page = 1
	}
	perPage := atoiDefault(q.Get("per_page"), application.DefaultPerPage)

	f := domain.ArticleFilter{
		Search:       q.Get("q"),
		ActiveOnly:   q.Get("include_inactive") != "true",
		HeadlineOnly: q.Get("headline") == "true",
		Limit:        perPage,
		Offset:       (page - 1) * perPage,
	}
	for _, raw := range strings.Split(q.Get("category"), ",") {
		if raw = strings.TrimSpace(raw); raw != "" {
			if id, err := strconv.ParseInt(raw, 10, 64); err == nil {
				f.CategoryIDs = append(f.CategoryIDs, id)
			}
		}
	}

	res, err := h.Service.ListArticles(r.Context(), f)
	if err != nil {
		render.Error(w, h.Log, middleware.RequestIDFrom(r.Context()), err)
		return
	}

	items := make([]articleDTO, 0, len(res.Items))
	for _, a := range res.Items {
		items = append(items, toDTO(a, false))
	}
	render.Paged(w, items, res.Total, page, perPage)
}

// BySlug handles GET /v1/articles/by-slug/{slug}.
//
// This replaces the page that loaded all 553 articles and slugified each one in
// JavaScript to find a single match.
func (h *Articles) BySlug(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	if slug == "" {
		render.BadRequest(w, "slug is required")
		return
	}
	a, err := h.Service.ArticleBySlug(r.Context(), slug, r.URL.Query().Get("count_view") == "true")
	if err != nil {
		render.Error(w, h.Log, middleware.RequestIDFrom(r.Context()), err)
		return
	}
	render.OK(w, toDTO(*a, true))
}

func (h *Articles) ByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		render.BadRequest(w, "id must be a number")
		return
	}
	a, err := h.Service.ArticleByID(r.Context(), id)
	if err != nil {
		render.Error(w, h.Log, middleware.RequestIDFrom(r.Context()), err)
		return
	}
	render.OK(w, toDTO(*a, true))
}

// categoryDTO exists for the same reason articleDTO does: returning a domain
// type directly leaks Go field names into the API (ID, Title, IsActive) and
// couples the wire format to internal naming.
type categoryDTO struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	IsActive bool   `json:"is_active"`
}

func (h *Articles) Categories(w http.ResponseWriter, r *http.Request) {
	cats, err := h.Service.Categories(r.Context())
	if err != nil {
		render.Error(w, h.Log, middleware.RequestIDFrom(r.Context()), err)
		return
	}
	out := make([]categoryDTO, 0, len(cats))
	for _, c := range cats {
		out = append(out, categoryDTO{ID: c.ID, Title: c.Title, IsActive: c.IsActive})
	}
	render.OK(w, out)
}

type articleInput struct {
	CategoryID  int64      `json:"category_id"`
	Title       string     `json:"title"`
	Content     string     `json:"content"`
	Picture     string     `json:"picture"`
	Editor      string     `json:"editor"`
	IsActive    *bool      `json:"is_active"`
	IsHeadline  *bool      `json:"is_headline"`
	PublishedAt *time.Time `json:"published_at"`
}

func (in articleInput) toDomain() *domain.Article {
	a := &domain.Article{
		CategoryID: in.CategoryID, Title: in.Title, Content: in.Content,
		Picture: in.Picture, Editor: in.Editor, IsActive: true,
	}
	if in.IsActive != nil {
		a.IsActive = *in.IsActive
	}
	if in.IsHeadline != nil {
		a.IsHeadline = *in.IsHeadline
	}
	if in.PublishedAt != nil {
		a.PublishedAt = *in.PublishedAt
	}
	return a
}

func (h *Articles) Create(w http.ResponseWriter, r *http.Request) {
	var in articleInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		render.BadRequest(w, "body must be valid JSON")
		return
	}
	a := in.toDomain()
	if err := h.Service.CreateArticle(r.Context(), a); err != nil {
		render.Error(w, h.Log, middleware.RequestIDFrom(r.Context()), err)
		return
	}
	render.Created(w, toDTO(*a, true))
}

func (h *Articles) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		render.BadRequest(w, "id must be a number")
		return
	}
	var in articleInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		render.BadRequest(w, "body must be valid JSON")
		return
	}
	a := in.toDomain()
	a.ID = id
	if err := h.Service.UpdateArticle(r.Context(), a); err != nil {
		render.Error(w, h.Log, middleware.RequestIDFrom(r.Context()), err)
		return
	}
	render.OK(w, toDTO(*a, true))
}

func (h *Articles) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		render.BadRequest(w, "id must be a number")
		return
	}
	if err := h.Service.DeleteArticle(r.Context(), id); err != nil {
		render.Error(w, h.Log, middleware.RequestIDFrom(r.Context()), err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func atoiDefault(s string, fallback int) int {
	if n, err := strconv.Atoi(s); err == nil {
		return n
	}
	return fallback
}

// excerpt strips tags and truncates on a word boundary.
func excerpt(html string, max int) string {
	var b strings.Builder
	depth := 0
	for _, r := range html {
		switch {
		case r == '<':
			depth++
		case r == '>':
			if depth > 0 {
				depth--
			}
		case depth == 0:
			b.WriteRune(r)
		}
	}
	text := strings.Join(strings.Fields(b.String()), " ")
	if len(text) <= max {
		return text
	}
	cut := text[:max]
	if i := strings.LastIndex(cut, " "); i > 0 {
		cut = cut[:i]
	}
	return cut + "…"
}
