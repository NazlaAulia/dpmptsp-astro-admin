package cache

import (
	"context"

	"dpmptsp/api/internal/domain"
)

// CachedSliderRepo caches the carousel list and retires the homepage on write:
// the slider is the first thing on the front page, so a stale one is the most
// visible kind of stale there is.
type CachedSliderRepo struct {
	inner domain.SliderRepository
	rdb   *Client
}

func NewCachedSliderRepo(inner domain.SliderRepository, rdb *Client) *CachedSliderRepo {
	return &CachedSliderRepo{inner: inner, rdb: rdb}
}

var _ domain.SliderRepository = (*CachedSliderRepo)(nil)

func (r *CachedSliderRepo) List(ctx context.Context) ([]domain.Slider, error) {
	key, ok := VersionedKey(ctx, r.rdb, ResourceContent, "sliders")
	if !ok {
		return r.inner.List(ctx)
	}
	if hit, found := GetJSON[[]domain.Slider](ctx, r.rdb, key); found {
		return hit, nil
	}
	v, err := r.inner.List(ctx)
	if err != nil {
		return v, err
	}
	SetJSON(ctx, r.rdb, key, v, TTLContent)
	return v, nil
}

// ByID backs the edit form, which must never show a stale value.
func (r *CachedSliderRepo) ByID(ctx context.Context, id int64) (*domain.Slider, error) {
	return r.inner.ByID(ctx, id)
}

func (r *CachedSliderRepo) invalidate(ctx context.Context, err error) error {
	if err != nil {
		return err
	}
	Invalidate(ctx, r.rdb, ResourceContent)
	Invalidate(ctx, r.rdb, ResourceHome)
	return nil
}

func (r *CachedSliderRepo) Create(ctx context.Context, s *domain.Slider) error {
	return r.invalidate(ctx, r.inner.Create(ctx, s))
}

func (r *CachedSliderRepo) Update(ctx context.Context, s *domain.Slider) error {
	return r.invalidate(ctx, r.inner.Update(ctx, s))
}

func (r *CachedSliderRepo) Delete(ctx context.Context, id int64) error {
	return r.invalidate(ctx, r.inner.Delete(ctx, id))
}
