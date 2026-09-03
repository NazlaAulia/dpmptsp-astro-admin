package cache

import (
	"context"
	"fmt"

	"dpmptsp/api/internal/domain"
)

// CachedPageRepo caches the composed page payloads.
//
// /v1/home is the highest-traffic endpoint and assembles ten sections, so it is
// the single best candidate for caching in the whole API.
type CachedPageRepo struct {
	inner domain.PageRepository
	rdb   *Client
}

func NewCachedPageRepo(inner domain.PageRepository, rdb *Client) *CachedPageRepo {
	return &CachedPageRepo{inner: inner, rdb: rdb}
}

var _ domain.PageRepository = (*CachedPageRepo)(nil)

func (r *CachedPageRepo) Home(ctx context.Context) (domain.HomeContent, error) {
	key, ok := VersionedKey(ctx, r.rdb, ResourceContent, "home")
	if !ok {
		return r.inner.Home(ctx)
	}
	if hit, found := GetJSON[domain.HomeContent](ctx, r.rdb, key); found {
		return hit, nil
	}
	v, err := r.inner.Home(ctx)
	if err != nil {
		return v, err
	}
	SetJSON(ctx, r.rdb, key, v, TTLContent)
	return v, nil
}

func (r *CachedPageRepo) About(ctx context.Context) (domain.AboutContentPage, error) {
	key, ok := VersionedKey(ctx, r.rdb, ResourceContent, "about")
	if !ok {
		return r.inner.About(ctx)
	}
	if hit, found := GetJSON[domain.AboutContentPage](ctx, r.rdb, key); found {
		return hit, nil
	}
	v, err := r.inner.About(ctx)
	if err != nil {
		return v, err
	}
	SetJSON(ctx, r.rdb, key, v, TTLContent)
	return v, nil
}

func (r *CachedPageRepo) ServiceStandards(ctx context.Context) ([]domain.ServiceStandard, error) {
	key, ok := VersionedKey(ctx, r.rdb, ResourceContent, "service-standards")
	if !ok {
		return r.inner.ServiceStandards(ctx)
	}
	if hit, found := GetJSON[[]domain.ServiceStandard](ctx, r.rdb, key); found {
		return hit, nil
	}
	v, err := r.inner.ServiceStandards(ctx)
	if err != nil {
		return v, err
	}
	SetJSON(ctx, r.rdb, key, v, TTLContent)
	return v, nil
}

func (r *CachedPageRepo) InfoSection(ctx context.Context, sectionID string) (*domain.InfoSection, error) {
	key, ok := VersionedKey(ctx, r.rdb, ResourceContent, "info-section", sectionID)
	if !ok {
		return r.inner.InfoSection(ctx, sectionID)
	}
	if hit, found := GetJSON[domain.InfoSection](ctx, r.rdb, key); found {
		return &hit, nil
	}
	v, err := r.inner.InfoSection(ctx, sectionID)
	if err != nil {
		return nil, err
	}
	SetJSON(ctx, r.rdb, key, *v, TTLContent)
	return v, nil
}

// Photos is paginated, so the offset and limit belong in the key.
func (r *CachedPageRepo) Photos(ctx context.Context, limit, offset int) (domain.Page[domain.Photo], error) {
	key, ok := VersionedKey(ctx, r.rdb, ResourceContent, "photos",
		fmt.Sprintf("limit=%d", limit), fmt.Sprintf("offset=%d", offset))
	if !ok {
		return r.inner.Photos(ctx, limit, offset)
	}
	if hit, found := GetJSON[domain.Page[domain.Photo]](ctx, r.rdb, key); found {
		return hit, nil
	}
	v, err := r.inner.Photos(ctx, limit, offset)
	if err != nil {
		return v, err
	}
	SetJSON(ctx, r.rdb, key, v, TTLContent)
	return v, nil
}
