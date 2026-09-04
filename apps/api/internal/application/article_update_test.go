package application

import (
	"context"
	"testing"
	"time"

	"dpmptsp/api/internal/domain"
)

// stubArticleRepo implements domain.ArticleRepository; only the two methods
// UpdateArticle actually calls do anything.
type stubArticleRepo struct {
	existing *domain.Article
	updated  *domain.Article
}

func (s *stubArticleRepo) List(context.Context, domain.ArticleFilter) (domain.Page[domain.Article], error) {
	return domain.Page[domain.Article]{}, nil
}
func (s *stubArticleRepo) BySlug(context.Context, string) (*domain.Article, error) {
	return nil, domain.ErrNotFound
}
func (s *stubArticleRepo) ByID(context.Context, int64) (*domain.Article, error) {
	return s.existing, nil
}
func (s *stubArticleRepo) Create(context.Context, *domain.Article) error { return nil }
func (s *stubArticleRepo) Update(_ context.Context, a *domain.Article) error {
	s.updated = a
	return nil
}
func (s *stubArticleRepo) Delete(context.Context, int64) error        { return nil }
func (s *stubArticleRepo) IncrementHits(context.Context, int64) error { return nil }

// An edit form that does not carry the publication date must not erase it.
func TestUpdateArticleKeepsFieldsTheCallerOmitted(t *testing.T) {
	published := time.Date(2024, 3, 9, 8, 0, 0, 0, time.UTC)
	repo := &stubArticleRepo{existing: &domain.Article{
		ID: 1, Title: "Judul", Slug: "judul", PublishedAt: published, Editor: "rina",
	}}
	svc := &ArticleService{articles: repo}

	in := &domain.Article{ID: 1, CategoryID: 1, Title: "Judul", Content: "isi"}
	if err := svc.UpdateArticle(context.Background(), in); err != nil {
		t.Fatal(err)
	}

	if !repo.updated.PublishedAt.Equal(published) {
		t.Errorf("published_at = %v, want the stored %v", repo.updated.PublishedAt, published)
	}
	if repo.updated.Editor != "rina" {
		t.Errorf("editor = %q, want the stored %q", repo.updated.Editor, "rina")
	}
	if repo.updated.Slug != "judul" {
		t.Errorf("slug = %q, want it unchanged", repo.updated.Slug)
	}
}

// A caller that does supply the fields still wins.
func TestUpdateArticleAcceptsSuppliedFields(t *testing.T) {
	repo := &stubArticleRepo{existing: &domain.Article{
		ID: 1, Title: "Judul", Slug: "judul",
		PublishedAt: time.Date(2024, 3, 9, 8, 0, 0, 0, time.UTC), Editor: "rina",
	}}
	svc := &ArticleService{articles: repo}

	want := time.Date(2026, 1, 2, 9, 0, 0, 0, time.UTC)
	in := &domain.Article{ID: 1, CategoryID: 1, Title: "Judul", Content: "isi", PublishedAt: want, Editor: "budi"}
	if err := svc.UpdateArticle(context.Background(), in); err != nil {
		t.Fatal(err)
	}

	if !repo.updated.PublishedAt.Equal(want) {
		t.Errorf("published_at = %v, want %v", repo.updated.PublishedAt, want)
	}
	if repo.updated.Editor != "budi" {
		t.Errorf("editor = %q, want budi", repo.updated.Editor)
	}
}
