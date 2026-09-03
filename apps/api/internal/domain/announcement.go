package domain

import "context"

// AnnouncementType selects how an announcement is displayed.
const (
	AnnouncementNotif = "notif"
	AnnouncementModal = "modal"
)

type Announcement struct {
	ID        int64
	Judul     string
	Deskripsi string
	Foto      string
	LinkURL   string
	IsActive  bool
	Urutan    int
	Tipe      string
	CreatedAt string
	UpdatedAt string
}

type AnnouncementFilter struct {
	// Tipe filters by display type when set.
	Tipe string
	// ActiveOnly restricts the result to published rows.
	ActiveOnly bool
}

type AnnouncementRepository interface {
	List(ctx context.Context, f AnnouncementFilter) ([]Announcement, error)
	ByID(ctx context.Context, id int64) (*Announcement, error)
	Create(ctx context.Context, a *Announcement) error
	Update(ctx context.Context, a *Announcement) error
	Delete(ctx context.Context, id int64) error
}
