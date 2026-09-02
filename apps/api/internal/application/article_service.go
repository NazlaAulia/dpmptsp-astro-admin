// Package application holds the use cases. It depends only on domain
// interfaces, so it can be tested without a database and cannot tell which
// engine — or which cache — is underneath.
package application

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"dpmptsp/api/internal/domain"
)

const (
	DefaultPerPage = 10
	MaxPerPage     = 100
)

type ArticleService struct {
	articles   domain.ArticleRepository
	categories domain.CategoryRepository
}

func NewArticleService(a domain.ArticleRepository, c domain.CategoryRepository) *ArticleService {
	return &ArticleService{articles: a, categories: c}
}

// ListArticles clamps paging rather than trusting the caller.
//
// The Astro pages currently derive OFFSET from a query string with no bounds
// check, so ?page=0 produces a negative offset and a SQL error, and a huge
// ?limit would let a caller ask for the entire table in one request.
func (s *ArticleService) ListArticles(ctx context.Context, f domain.ArticleFilter) (domain.Page[domain.Article], error) {
	if f.Limit <= 0 {
		f.Limit = DefaultPerPage
	}
	if f.Limit > MaxPerPage {
		f.Limit = MaxPerPage
	}
	if f.Offset < 0 {
		f.Offset = 0
	}
	f.Search = strings.TrimSpace(f.Search)
	return s.articles.List(ctx, f)
}

// ArticleBySlug returns one article and counts the view.
func (s *ArticleService) ArticleBySlug(ctx context.Context, slug string, countView bool) (*domain.Article, error) {
	a, err := s.articles.BySlug(ctx, strings.TrimSpace(slug))
	if err != nil {
		return nil, err
	}
	if countView {
		// A failed counter must not fail the page.
		_ = s.articles.IncrementHits(ctx, a.ID)
	}
	return a, nil
}

func (s *ArticleService) ArticleByID(ctx context.Context, id int64) (*domain.Article, error) {
	return s.articles.ByID(ctx, id)
}

func (s *ArticleService) Categories(ctx context.Context) ([]domain.Category, error) {
	return s.categories.All(ctx)
}

// CreateArticle validates, assigns a unique slug, and stores.
func (s *ArticleService) CreateArticle(ctx context.Context, a *domain.Article) error {
	if err := validate(a); err != nil {
		return err
	}
	if a.PublishedAt.IsZero() {
		a.PublishedAt = time.Now()
	}
	slug, err := s.uniqueSlug(ctx, Slugify(a.Title), 0)
	if err != nil {
		return err
	}
	a.Slug = slug
	return s.articles.Create(ctx, a)
}

func (s *ArticleService) UpdateArticle(ctx context.Context, a *domain.Article) error {
	if err := validate(a); err != nil {
		return err
	}
	existing, err := s.articles.ByID(ctx, a.ID)
	if err != nil {
		return err
	}
	// Only re-slug when the title actually changed: a slug is a URL, and
	// changing it silently breaks every existing link to the article.
	if existing.Title != a.Title {
		slug, err := s.uniqueSlug(ctx, Slugify(a.Title), a.ID)
		if err != nil {
			return err
		}
		a.Slug = slug
	} else {
		a.Slug = existing.Slug
	}
	return s.articles.Update(ctx, a)
}

func (s *ArticleService) DeleteArticle(ctx context.Context, id int64) error {
	return s.articles.Delete(ctx, id)
}

func validate(a *domain.Article) error {
	if strings.TrimSpace(a.Title) == "" {
		return fmt.Errorf("%w: title is required", domain.ErrInvalid)
	}
	if a.CategoryID <= 0 {
		return fmt.Errorf("%w: category is required", domain.ErrInvalid)
	}
	return nil
}

// uniqueSlug appends -2, -3 … until the slug is free, ignoring the article
// being updated.
func (s *ArticleService) uniqueSlug(ctx context.Context, base string, ignoreID int64) (string, error) {
	if base == "" {
		base = "artikel"
	}
	candidate := base
	for i := 2; i < 500; i++ {
		existing, err := s.articles.BySlug(ctx, candidate)
		if err == domain.ErrNotFound {
			return candidate, nil
		}
		if err != nil {
			return "", err
		}
		if existing.ID == ignoreID {
			return candidate, nil
		}
		candidate = fmt.Sprintf("%s-%d", base, i)
	}
	return "", fmt.Errorf("%w: could not find a free slug for %q", domain.ErrConflict, base)
}

var nonSlug = regexp.MustCompile(`[^a-z0-9]+`)

// Slugify reproduces the behaviour of apps/web/src/utils/slugify.js EXACTLY,
// including two quirks, because article URLs already derive from it and a
// "better" implementation would 404 every existing link:
//
//   - an en dash surrounded by spaces yields a DOUBLE hyphen, because both
//     spaces become hyphens and the dash itself is then stripped
//   - accented letters are DELETED, not transliterated: É disappears rather
//     than folding to E
//
// The JS is `.toLowerCase().trim().replace(/\s+/g,"-").replace(/[^\w\-]/g,"")`,
// and \w includes the underscore, so underscores survive.
func Slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = regexp.MustCompile(`\s+`).ReplaceAllString(s, "-")
	// \w is [A-Za-z0-9_]; everything else goes, hyphens included in the keep set.
	s = regexp.MustCompile(`[^a-z0-9_\-]`).ReplaceAllString(s, "")
	return s
}
