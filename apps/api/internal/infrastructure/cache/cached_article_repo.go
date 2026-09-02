package cache

import (
	"context"

	"dpmptsp/api/internal/domain"
)

// CachedArticleRepo is the decorator SPEC.md §6 describes: cache-aside in front
// of the real repository, wired in main.go, invisible to everything above it.
//
// The application layer receives a domain.ArticleRepository and cannot tell a
// cache is here, which is what lets the whole thing be removed by changing one
// line of wiring.
type CachedArticleRepo struct {
	inner domain.ArticleRepository
	rdb   *Client
}

func NewCachedArticleRepo(inner domain.ArticleRepository, rdb *Client) *CachedArticleRepo {
	return &CachedArticleRepo{inner: inner, rdb: rdb}
}

var _ domain.ArticleRepository = (*CachedArticleRepo)(nil)

func (r *CachedArticleRepo) List(ctx context.Context, f domain.ArticleFilter) (domain.Page[domain.Article], error) {
	// Search is deliberately NOT cached. The key space is unbounded — every
	// distinct query string is a new key — and the hit rate is near zero. The
	// fix for slow search is the full-text index, not a cache.
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

// ByID is not cached. It is only reached from the admin panel, and an editor
// who saves a change and does not see it is a worse bug than a slow page.
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
	// Read the row first so its slug key can be dropped: after the delete there
	// is nothing left to tell us which key to remove, and the stale entry would
	// serve a deleted article until its TTL expired.
	existing, _ := r.inner.ByID(ctx, id)
	if err := r.inner.Delete(ctx, id); err != nil {
		return err
	}
	r.invalidate(ctx, existing)
	return nil
}

// IncrementHits does NOT invalidate. A view counter changes on nearly every
// request, and bumping the version each time would make the list cache useless.
// The counter being a few minutes stale is the intended trade.
func (r *CachedArticleRepo) IncrementHits(ctx context.Context, id int64) error {
	return r.inner.IncrementHits(ctx, id)
}

// invalidate bumps the list counter and drops the one entity key.
//
// Note what is absent: no SCAN, no KEYS, no pattern delete. Every list key
// built from the previous version simply stops being addressable and expires
// on its own TTL (CLAUDE.md rule 7).
func (r *CachedArticleRepo) invalidate(ctx context.Context, a *domain.Article) {
	Invalidate(ctx, r.rdb, ResourceArticles)
	if a != nil && a.Slug != "" {
		Del(ctx, r.rdb, EntityKey(ResourceArticles, "slug:"+a.Slug))
	}
}

// CachedSiteRepo caches the branding and menu.
//
// This is the highest-value entry in the whole cache: the header renders on
// every page and its data changes perhaps monthly, so it should be a near
// permanent hit with a long TTL.
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

// InvalidateChrome is called by whatever edits the menu or the settings.
func InvalidateChrome(ctx context.Context, rdb *Client) {
	Invalidate(ctx, rdb, ResourceChrome)
}
