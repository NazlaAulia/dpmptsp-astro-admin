// Package routes wires handlers and middleware into the HTTP mux.
package routes

import (
	"log/slog"
	"net/http"
	"strings"

	"gorm.io/gorm"

	"dpmptsp/api/internal/application"
	"dpmptsp/api/internal/config"
	"dpmptsp/api/internal/http/handlers"
	"dpmptsp/api/internal/http/middleware"
	"dpmptsp/api/internal/infrastructure/cache"
	"dpmptsp/api/internal/infrastructure/database"
	"dpmptsp/api/internal/infrastructure/database/dialect"
	"dpmptsp/api/internal/infrastructure/storage"
)

func New(cfg *config.Config, log *slog.Logger, db *gorm.DB, dia dialect.Dialect, rdb *cache.Client, files *storage.Manager) http.Handler {
	mux := http.NewServeMux()

	health := &handlers.Health{DB: db, Redis: rdb, Engine: string(cfg.Engine()), Files: files}
	mux.HandleFunc("GET /healthz", health.Live)
	mux.HandleFunc("GET /readyz", health.Ready)

	// Articles. Wired here rather than in a framework: net/http's own router
	// handles method-and-pattern matching since 1.22, so a dependency would
	// buy nothing.
	// Cache-aside decorators, wired here and nowhere else. The
	// application layer receives domain interfaces and cannot tell a cache is
	// present, so removing it is a one-line change.
	articleRepo := cache.NewCachedArticleRepo(database.NewArticleRepo(db, dia), rdb)
	siteRepo := cache.NewCachedSiteRepo(database.NewSiteRepo(db), rdb)

	articles := &handlers.Articles{
		Service: application.NewArticleService(
			articleRepo,
			database.NewCategoryRepo(db),
		),
		Log: log,
	}
	mux.HandleFunc("GET /v1/articles", articles.List)
	mux.HandleFunc("GET /v1/articles/by-slug/{slug}", articles.BySlug)
	mux.HandleFunc("GET /v1/articles/{id}", articles.ByID)
	mux.HandleFunc("POST /v1/articles", articles.Create)
	mux.HandleFunc("PUT /v1/articles/{id}", articles.Update)
	mux.HandleFunc("DELETE /v1/articles/{id}", articles.Delete)
	mux.HandleFunc("GET /v1/categories", articles.Categories)

	// Site chrome: branding and the header menu, already nested.
	auth := &handlers.Auth{
		Service: application.NewAuthService(database.NewUserRepo(db)),
		Log:     log,
	}
	mux.HandleFunc("POST /v1/auth/login", auth.Login)

	uploads := &handlers.Uploads{Files: files, Log: log}
	mux.HandleFunc("POST /v1/uploads", uploads.Create)

	site := &handlers.Site{Repo: siteRepo, Log: log}
	mux.HandleFunc("GET /v1/site/chrome", site.Chrome)

	// Content lists. All read-only and rarely written; the admin panel still
	// writes them directly until those pages are ported too.
	content := &handlers.Content{Repo: database.NewContentRepo(db), Log: log}
	mux.HandleFunc("GET /v1/regulations", content.Regulations)
	mux.HandleFunc("GET /v1/public-apps", content.PublicApps)
	mux.HandleFunc("GET /v1/videos", content.Videos)
	mux.HandleFunc("GET /v1/performance-docs", content.PerformanceDocs)
	mux.HandleFunc("GET /v1/service-locations", content.ServiceLocations)
	mux.HandleFunc("GET /v1/ppid", content.PPID)

	// Admin writes.
	mux.HandleFunc("GET /v1/innovations", content.Innovations)
	mux.HandleFunc("GET /v1/innovations/{id}", content.Innovation)
	mux.HandleFunc("POST /v1/innovations", content.CreateInnovation)
	mux.HandleFunc("PUT /v1/innovations/{id}", content.UpdateInnovation)
	mux.HandleFunc("DELETE /v1/innovations/{id}", content.DeleteInnovation)

	mux.HandleFunc("GET /v1/performance-docs/{id}", content.PerformanceDoc)
	mux.HandleFunc("POST /v1/performance-docs", content.CreatePerformanceDoc)
	mux.HandleFunc("PUT /v1/performance-docs/{id}", content.UpdatePerformanceDoc)
	mux.HandleFunc("DELETE /v1/performance-docs/{id}", content.DeletePerformanceDoc)

	mux.HandleFunc("GET /v1/service-locations/{id}", content.ServiceLocation)
	mux.HandleFunc("POST /v1/service-locations", content.CreateServiceLocation)
	mux.HandleFunc("PUT /v1/service-locations/{id}", content.UpdateServiceLocation)
	mux.HandleFunc("DELETE /v1/service-locations/{id}", content.DeleteServiceLocation)

	mux.HandleFunc("GET /v1/about-contents", content.AboutContents)
	mux.HandleFunc("GET /v1/about-contents/{id}", content.AboutContent)
	mux.HandleFunc("POST /v1/about-contents", content.CreateAboutContent)
	mux.HandleFunc("PUT /v1/about-contents/{id}", content.UpdateAboutContent)
	mux.HandleFunc("DELETE /v1/about-contents/{id}", content.DeleteAboutContent)

	// Composite page payloads.
	contentRepo := database.NewContentRepo(db)
	pages := &handlers.Pages{
		Repo: database.NewPageRepo(db, contentRepo, database.NewArticleRepo(db, dia)),
		Log:  log,
	}
	mux.HandleFunc("GET /v1/home", pages.Home)
	mux.HandleFunc("GET /v1/about", pages.About)
	mux.HandleFunc("GET /v1/service-standards", pages.ServiceStandards)
	mux.HandleFunc("GET /v1/info-sections/{sectionId}", pages.InfoSection)
	mux.HandleFunc("GET /v1/gallery/photos", pages.Photos)

	announcements := &handlers.Announcements{Repo: database.NewAnnouncementRepo(db), Log: log}
	mux.HandleFunc("GET /v1/announcements", announcements.List)
	mux.HandleFunc("GET /v1/announcements/{id}", announcements.ByID)
	mux.HandleFunc("POST /v1/announcements", announcements.Create)
	mux.HandleFunc("PUT /v1/announcements/{id}", announcements.Update)
	mux.HandleFunc("DELETE /v1/announcements/{id}", announcements.Delete)

	// Health endpoints stay reachable without the service key so that container
	// orchestration can probe them.
	skipAuth := func(r *http.Request) bool {
		return strings.HasPrefix(r.URL.Path, "/healthz") ||
			strings.HasPrefix(r.URL.Path, "/readyz")
	}

	return middleware.Chain(
		mux,
		middleware.RequestID,
		middleware.Recover(log),
		middleware.Logging(log),
		middleware.ServiceAuth(cfg.ServiceKey, skipAuth),
	)
}
