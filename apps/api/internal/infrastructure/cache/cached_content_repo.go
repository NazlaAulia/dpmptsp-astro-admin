package cache

import (
	"context"

	"dpmptsp/api/internal/domain"
)

// CachedContentRepo caches the public content lists and retires them on write.
//
// Every list shares the content version counter, so one INCR after any write
// retires all of them — including the composed home and about payloads, which
// embed the same data.
type CachedContentRepo struct {
	inner domain.ContentRepository
	rdb   *Client
}

func NewCachedContentRepo(inner domain.ContentRepository, rdb *Client) *CachedContentRepo {
	return &CachedContentRepo{inner: inner, rdb: rdb}
}

var _ domain.ContentRepository = (*CachedContentRepo)(nil)

// cachedList reads a list through the cache, falling straight through when
// Redis is unavailable.
func cachedList[T any](ctx context.Context, rdb *Client, name string, load func() ([]T, error)) ([]T, error) {
	key, ok := VersionedKey(ctx, rdb, ResourceContent, name)
	if !ok {
		return load()
	}
	if hit, found := GetJSON[[]T](ctx, rdb, key); found {
		return hit, nil
	}
	v, err := load()
	if err != nil {
		return v, err
	}
	SetJSON(ctx, rdb, key, v, TTLContent)
	return v, nil
}

func (r *CachedContentRepo) Regulations(ctx context.Context) ([]domain.Regulation, error) {
	return cachedList(ctx, r.rdb, "regulations", func() ([]domain.Regulation, error) {
		return r.inner.Regulations(ctx)
	})
}

func (r *CachedContentRepo) PublicApps(ctx context.Context) ([]domain.PublicApp, error) {
	return cachedList(ctx, r.rdb, "public-apps", func() ([]domain.PublicApp, error) {
		return r.inner.PublicApps(ctx)
	})
}

func (r *CachedContentRepo) Videos(ctx context.Context) ([]domain.Video, error) {
	return cachedList(ctx, r.rdb, "videos", func() ([]domain.Video, error) {
		return r.inner.Videos(ctx)
	})
}

func (r *CachedContentRepo) PerformanceDocs(ctx context.Context) ([]domain.PerformanceDoc, error) {
	return cachedList(ctx, r.rdb, "performance-docs", func() ([]domain.PerformanceDoc, error) {
		return r.inner.PerformanceDocs(ctx)
	})
}

func (r *CachedContentRepo) ServiceLocations(ctx context.Context) ([]domain.ServiceLocation, error) {
	return cachedList(ctx, r.rdb, "service-locations", func() ([]domain.ServiceLocation, error) {
		return r.inner.ServiceLocations(ctx)
	})
}

func (r *CachedContentRepo) PPID(ctx context.Context) ([]domain.PPIDCategory, error) {
	return cachedList(ctx, r.rdb, "ppid", func() ([]domain.PPIDCategory, error) {
		return r.inner.PPID(ctx)
	})
}

func (r *CachedContentRepo) Innovations(ctx context.Context) ([]domain.Innovation, error) {
	return cachedList(ctx, r.rdb, "innovations", func() ([]domain.Innovation, error) {
		return r.inner.Innovations(ctx)
	})
}

func (r *CachedContentRepo) AboutContents(ctx context.Context) ([]domain.AboutContent, error) {
	return cachedList(ctx, r.rdb, "about-contents", func() ([]domain.AboutContent, error) {
		return r.inner.AboutContents(ctx)
	})
}

// Single-row reads are not cached: they are reached from admin edit screens,
// where showing an editor a stale copy of what they just saved is worse than
// the extra query.

func (r *CachedContentRepo) Innovation(ctx context.Context, id int64) (*domain.Innovation, error) {
	return r.inner.Innovation(ctx, id)
}

func (r *CachedContentRepo) PerformanceDoc(ctx context.Context, id int64) (*domain.PerformanceDoc, error) {
	return r.inner.PerformanceDoc(ctx, id)
}

func (r *CachedContentRepo) ServiceLocation(ctx context.Context, id int64) (*domain.ServiceLocation, error) {
	return r.inner.ServiceLocation(ctx, id)
}

func (r *CachedContentRepo) AboutContentByID(ctx context.Context, id int64) (*domain.AboutContent, error) {
	return r.inner.AboutContentByID(ctx, id)
}

// Writes: perform, then retire every content list at once.

func (r *CachedContentRepo) invalidate(ctx context.Context, err error) error {
	if err != nil {
		return err
	}
	Invalidate(ctx, r.rdb, ResourceContent)
	// The homepage embeds these lists as well.
	Invalidate(ctx, r.rdb, ResourceHome)
	return nil
}

func (r *CachedContentRepo) CreateInnovation(ctx context.Context, v *domain.Innovation) error {
	return r.invalidate(ctx, r.inner.CreateInnovation(ctx, v))
}
func (r *CachedContentRepo) UpdateInnovation(ctx context.Context, v *domain.Innovation) error {
	return r.invalidate(ctx, r.inner.UpdateInnovation(ctx, v))
}
func (r *CachedContentRepo) DeleteInnovation(ctx context.Context, id int64) error {
	return r.invalidate(ctx, r.inner.DeleteInnovation(ctx, id))
}

func (r *CachedContentRepo) CreatePerformanceDoc(ctx context.Context, v *domain.PerformanceDoc) error {
	return r.invalidate(ctx, r.inner.CreatePerformanceDoc(ctx, v))
}
func (r *CachedContentRepo) UpdatePerformanceDoc(ctx context.Context, v *domain.PerformanceDoc) error {
	return r.invalidate(ctx, r.inner.UpdatePerformanceDoc(ctx, v))
}
func (r *CachedContentRepo) DeletePerformanceDoc(ctx context.Context, id int64) error {
	return r.invalidate(ctx, r.inner.DeletePerformanceDoc(ctx, id))
}

func (r *CachedContentRepo) CreateServiceLocation(ctx context.Context, v *domain.ServiceLocation) error {
	return r.invalidate(ctx, r.inner.CreateServiceLocation(ctx, v))
}
func (r *CachedContentRepo) UpdateServiceLocation(ctx context.Context, v *domain.ServiceLocation) error {
	return r.invalidate(ctx, r.inner.UpdateServiceLocation(ctx, v))
}
func (r *CachedContentRepo) DeleteServiceLocation(ctx context.Context, id int64) error {
	return r.invalidate(ctx, r.inner.DeleteServiceLocation(ctx, id))
}

func (r *CachedContentRepo) CreateAboutContent(ctx context.Context, v *domain.AboutContent) error {
	return r.invalidate(ctx, r.inner.CreateAboutContent(ctx, v))
}
func (r *CachedContentRepo) UpdateAboutContent(ctx context.Context, v *domain.AboutContent) error {
	return r.invalidate(ctx, r.inner.UpdateAboutContent(ctx, v))
}
func (r *CachedContentRepo) DeleteAboutContent(ctx context.Context, id int64) error {
	return r.invalidate(ctx, r.inner.DeleteAboutContent(ctx, id))
}
