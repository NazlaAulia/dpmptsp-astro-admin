package cache

import (
	"context"
	"fmt"

	"dpmptsp/api/internal/domain"
)

// CachedAnnouncementRepo caches announcement reads and bumps the content
// counter on every write, which also retires the cached home payload — the
// homepage renders announcements.
type CachedAnnouncementRepo struct {
	inner domain.AnnouncementRepository
	rdb   *Client
}

func NewCachedAnnouncementRepo(inner domain.AnnouncementRepository, rdb *Client) *CachedAnnouncementRepo {
	return &CachedAnnouncementRepo{inner: inner, rdb: rdb}
}

var _ domain.AnnouncementRepository = (*CachedAnnouncementRepo)(nil)

func (r *CachedAnnouncementRepo) List(ctx context.Context, f domain.AnnouncementFilter) ([]domain.Announcement, error) {
	key, ok := VersionedKey(ctx, r.rdb, ResourceContent, "announcements",
		fmt.Sprintf("tipe=%s", f.Tipe), fmt.Sprintf("active=%t", f.ActiveOnly))
	if !ok {
		return r.inner.List(ctx, f)
	}
	if hit, found := GetJSON[[]domain.Announcement](ctx, r.rdb, key); found {
		return hit, nil
	}
	v, err := r.inner.List(ctx, f)
	if err != nil {
		return v, err
	}
	SetJSON(ctx, r.rdb, key, v, TTLContent)
	return v, nil
}

// ByID is not cached: it is reached from the admin edit screen, where a stale
// read would show an editor their own change missing.
func (r *CachedAnnouncementRepo) ByID(ctx context.Context, id int64) (*domain.Announcement, error) {
	return r.inner.ByID(ctx, id)
}

func (r *CachedAnnouncementRepo) Create(ctx context.Context, a *domain.Announcement) error {
	if err := r.inner.Create(ctx, a); err != nil {
		return err
	}
	Invalidate(ctx, r.rdb, ResourceContent)
	// Announcements are the homepage notification strip and modal.
	Invalidate(ctx, r.rdb, ResourceHome)
	return nil
}

func (r *CachedAnnouncementRepo) Update(ctx context.Context, a *domain.Announcement) error {
	if err := r.inner.Update(ctx, a); err != nil {
		return err
	}
	Invalidate(ctx, r.rdb, ResourceContent)
	// Announcements are the homepage notification strip and modal.
	Invalidate(ctx, r.rdb, ResourceHome)
	return nil
}

func (r *CachedAnnouncementRepo) Delete(ctx context.Context, id int64) error {
	if err := r.inner.Delete(ctx, id); err != nil {
		return err
	}
	Invalidate(ctx, r.rdb, ResourceContent)
	// Announcements are the homepage notification strip and modal.
	Invalidate(ctx, r.rdb, ResourceHome)
	return nil
}
