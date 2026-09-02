package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"dpmptsp/api/internal/domain"
	"dpmptsp/api/internal/infrastructure/database/dialect"
)

// ArticleRepo is the only implementation, for both engines. The dialect covers
// the one query that has no shared syntax.
type ArticleRepo struct {
	db  *gorm.DB
	dia dialect.Dialect
}

func NewArticleRepo(db *gorm.DB, dia dialect.Dialect) *ArticleRepo {
	return &ArticleRepo{db: db, dia: dia}
}

var _ domain.ArticleRepository = (*ArticleRepo)(nil)

func (r *ArticleRepo) query(ctx context.Context, f domain.ArticleFilter) *gorm.DB {
	q := r.db.WithContext(ctx).Model(&articleModel{})

	if len(f.CategoryIDs) > 0 {
		q = q.Where("id_category IN ?", f.CategoryIDs)
	}
	if f.ActiveOnly {
		q = q.Where("is_active = ?", true)
	}
	if f.HeadlineOnly {
		q = q.Where("is_headline = ?", true)
	}
	if f.Search != "" {
		cond, arg := r.dia.FullTextCondition([]string{"title", "content"}, f.Search)
		q = q.Where(cond, arg)
	}
	return q
}

func (r *ArticleRepo) List(ctx context.Context, f domain.ArticleFilter) (domain.Page[domain.Article], error) {
	var out domain.Page[domain.Article]

	if err := r.query(ctx, f).Count(&out.Total).Error; err != nil {
		return out, fmt.Errorf("count articles: %w", err)
	}
	if out.Total == 0 {
		out.Items = []domain.Article{}
		return out, nil
	}

	var rows []articleModel
	q := r.query(ctx, f).
		Preload("Category").
		Order("published_at DESC").
		Order("id_post DESC") // stable tiebreak, so paging cannot repeat a row
	if f.Limit > 0 {
		q = q.Limit(f.Limit)
	}
	if f.Offset > 0 {
		q = q.Offset(f.Offset)
	}
	if err := q.Find(&rows).Error; err != nil {
		return out, fmt.Errorf("list articles: %w", err)
	}

	out.Items = make([]domain.Article, 0, len(rows))
	for _, m := range rows {
		out.Items = append(out.Items, m.toDomain())
	}
	return out, nil
}

func (r *ArticleRepo) BySlug(ctx context.Context, slug string) (*domain.Article, error) {
	return r.one(ctx, "slug = ?", slug)
}

func (r *ArticleRepo) ByID(ctx context.Context, id int64) (*domain.Article, error) {
	return r.one(ctx, "id_post = ?", id)
}

func (r *ArticleRepo) one(ctx context.Context, where string, arg any) (*domain.Article, error) {
	var m articleModel
	err := r.db.WithContext(ctx).Model(&articleModel{}).
		Preload("Category").Where(where, arg).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find article: %w", err)
	}
	a := m.toDomain()
	return &a, nil
}

func (r *ArticleRepo) Create(ctx context.Context, a *domain.Article) error {
	m := articleFromDomain(a)
	m.CreatedAt = time.Now()
	m.UpdatedAt = m.CreatedAt
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return fmt.Errorf("create article: %w", err)
	}
	a.ID = m.ID
	return nil
}

func (r *ArticleRepo) Update(ctx context.Context, a *domain.Article) error {
	m := articleFromDomain(a)
	m.UpdatedAt = time.Now()
	res := r.db.WithContext(ctx).Model(&articleModel{}).
		Where("id_post = ?", a.ID).
		// Explicit column list. Save() would write every field including
		// created_at and hits, silently resetting the view counter.
		Select("id_category", "slug", "title", "content", "ref_content",
			"seotitle", "tag", "published_at", "editor", "is_active",
			"is_headline", "picture", "title_en", "content_en",
			"ref_content_en", "updated_at").
		Updates(m)
	if res.Error != nil {
		return fmt.Errorf("update article: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *ArticleRepo) Delete(ctx context.Context, id int64) error {
	res := r.db.WithContext(ctx).Where("id_post = ?", id).Delete(&articleModel{})
	if res.Error != nil {
		return fmt.Errorf("delete article: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// IncrementHits is an atomic UPDATE rather than a read-modify-write: two
// concurrent readers would otherwise both write the same value and lose a view.
func (r *ArticleRepo) IncrementHits(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&articleModel{}).
		Where("id_post = ?", id).
		UpdateColumn("hits", gorm.Expr("hits + 1")).Error
}

// CategoryRepo lists article categories.
type CategoryRepo struct{ db *gorm.DB }

func NewCategoryRepo(db *gorm.DB) *CategoryRepo { return &CategoryRepo{db: db} }

var _ domain.CategoryRepository = (*CategoryRepo)(nil)

func (r *CategoryRepo) All(ctx context.Context) ([]domain.Category, error) {
	var rows []categoryModel
	if err := r.db.WithContext(ctx).Order("id_category").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	out := make([]domain.Category, 0, len(rows))
	for _, m := range rows {
		out = append(out, m.toDomain())
	}
	return out, nil
}
