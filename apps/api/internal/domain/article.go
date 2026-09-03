package domain

import (
	"context"
	"time"
)

// Article is a news item or article. It maps to the legacy `post` table, whose
// column names are kept so the existing data needs no rewriting.
type Article struct {
	ID          int64
	CategoryID  int64
	Slug        string
	Title       string
	Content     string
	RefContent  string
	SEOTitle    string
	Tag         string
	PublishedAt time.Time
	Editor      string
	IsActive    bool
	IsHeadline  bool
	Picture     string
	Hits        int64

	// English variants. The bilingual effort was abandoned — no page renders
	// them today — but the columns hold data, so they are carried rather than
	// silently dropped.
	TitleEN      *string
	ContentEN    *string
	RefContentEN *string

	// Category is filled by the repository when it joins.
	Category *Category
}

type Category struct {
	ID       int64
	Title    string
	IsActive bool
}

// ArticleFilter describes a list query.
type ArticleFilter struct {
	CategoryIDs  []int64
	Search       string
	ActiveOnly   bool
	HeadlineOnly bool

	Limit  int
	Offset int
}

// Page is one page of results plus the total, so a caller can render
// pagination without a second round trip.
type Page[T any] struct {
	Items []T
	Total int64
}

type ArticleRepository interface {
	List(ctx context.Context, f ArticleFilter) (Page[Article], error)
	BySlug(ctx context.Context, slug string) (*Article, error)
	ByID(ctx context.Context, id int64) (*Article, error)
	Create(ctx context.Context, a *Article) error
	Update(ctx context.Context, a *Article) error
	Delete(ctx context.Context, id int64) error
	// IncrementHits is separate from Update: it is a hot path and must not
	// read-modify-write a whole row.
	IncrementHits(ctx context.Context, id int64) error
}

type CategoryRepository interface {
	All(ctx context.Context) ([]Category, error)
}
