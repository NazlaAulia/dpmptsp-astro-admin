package cache

import (
	"context"

	"dpmptsp/api/internal/domain"
)

// CachedArticleRepo is a cache-aside decorator over a domain.ArticleRepository.
// It is wired in routes.go and is transparent to the application layer.
type CachedArticleRepo struct {
	inner domain.ArticleRepository
	rdb   *Client
}

func NewCachedArticleRepo(inner domain.ArticleRepository, rdb *Client) *CachedArticleRepo {
	return &CachedArticleRepo{inner: inner, rdb: rdb}
}

var _ domain.ArticleRepository = (*CachedArticleRepo)(nil)

func (r *CachedArticleRepo) List(ctx context.Context, f domain.ArticleFilter) (domain.Page[domain.Article], error) {
	// Search is not cached: the key space is unbounded and the hit rate is
	// negligible.
	if f.Search != "" {
		return r.inner.List(ctx, f)
	}

	key, ok := ArticleListKey(ctx, r.rdb, f)
	if !ok {
		return r.inner.List(ctx, f)
	}

	if hit, found := GetJSON[domain.Page[domain.Article]](ctx, r.rdb, key); found {
		return hit, nil
	}

	page, err := r.inner.List(ctx, f)
	if err != nil {
		return page, err
	}
	SetJSON(ctx, r.rdb, key, page, TTLArticles)
	return page, nil
}

func (r *CachedArticleRepo) BySlug(ctx context.Context, slug string) (*domain.Article, error) {
	key := EntityKey(ResourceArticles, "slug:"+slug)
	if hit, found := GetJSON[domain.Article](ctx, r.rdb, key); found {
		return &hit, nil
	}
	a, err := r.inner.BySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	SetJSON(ctx, r.rdb, key, *a, TTLDetail)
	return a, nil
}

// ByID is not cached; it serves admin reads, which must be current.
func (r *CachedArticleRepo) ByID(ctx context.Context, id int64) (*domain.Article, error) {
	return r.inner.ByID(ctx, id)
}

func (r *CachedArticleRepo) Create(ctx context.Context, a *domain.Article) error {
	if err := r.inner.Create(ctx, a); err != nil {
		return err
	}
	r.invalidate(ctx, a)
	return nil
}

func (r *CachedArticleRepo) Update(ctx context.Context, a *domain.Article) error {
	if err := r.inner.Update(ctx, a); err != nil {
		return err
	}
	r.invalidate(ctx, a)
	return nil
}

func (r *CachedArticleRepo) Delete(ctx context.Context, id int64) error {
	// Read the row first so its slug key can be dropped after deletion.
	existing, _ := r.inner.ByID(ctx, id)
	if err := r.inner.Delete(ctx, id); err != nil {
		return err
	}
	r.invalidate(ctx, existing)
	return nil
}

// IncrementHits does not invalidate: the counter changes on nearly every
// request and may lag by one TTL.
func (r *CachedArticleRepo) IncrementHits(ctx context.Context, id int64) error {
	return r.inner.IncrementHits(ctx, id)
}

// invalidate bumps the list version counter and drops the entity key. List
// keys built from the previous version expire on their own TTL; no pattern
// delete is performed.
func (r *CachedArticleRepo) invalidate(ctx context.Context, a *domain.Article) {
	Invalidate(ctx, r.rdb, ResourceArticles)
	if a != nil && a.Slug != "" {
		Del(ctx, r.rdb, EntityKey(ResourceArticles, "slug:"+a.Slug))
	}
}

// CachedSiteRepo caches site branding and navigation, which are read on every
// page render and written rarely.
type CachedSiteRepo struct {
	inner domain.SiteRepository
	rdb   *Client
}

func NewCachedSiteRepo(inner domain.SiteRepository, rdb *Client) *CachedSiteRepo {
	return &CachedSiteRepo{inner: inner, rdb: rdb}
}

var _ domain.SiteRepository = (*CachedSiteRepo)(nil)

func (r *CachedSiteRepo) Settings(ctx context.Context) (domain.SiteSettings, error) {
	key, ok := VersionedKey(ctx, r.rdb, ResourceChrome, "settings")
	if !ok {
		return r.inner.Settings(ctx)
	}
	if hit, found := GetJSON[domain.SiteSettings](ctx, r.rdb, key); found {
		return hit, nil
	}
	s, err := r.inner.Settings(ctx)
	if err != nil {
		return s, err
	}
	SetJSON(ctx, r.rdb, key, s, TTLChrome)
	return s, nil
}

type menuPair struct {
	Nav     []domain.MenuNode
	Contact []domain.MenuNode
}

func (r *CachedSiteRepo) MenuTree(ctx context.Context) ([]domain.MenuNode, []domain.MenuNode, error) {
	key, ok := VersionedKey(ctx, r.rdb, ResourceChrome, "menu")
	if !ok {
		return r.inner.MenuTree(ctx)
	}
	if hit, found := GetJSON[menuPair](ctx, r.rdb, key); found {
		return hit.Nav, hit.Contact, nil
	}
	nav, contact, err := r.inner.MenuTree(ctx)
	if err != nil {
		return nil, nil, err
	}
	SetJSON(ctx, r.rdb, key, menuPair{Nav: nav, Contact: contact}, TTLChrome)
	return nav, contact, nil
}

// InvalidateChrome retires cached branding and navigation.
func InvalidateChrome(ctx context.Context, rdb *Client) {
	Invalidate(ctx, rdb, ResourceChrome)
}
