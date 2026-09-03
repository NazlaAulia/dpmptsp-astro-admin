package domain

import (
	"context"
	"time"
)

// ContactMessage is a message or complaint submitted from the public site.
type ContactMessage struct {
	ID         int64
	TicketCode string
	Nama       string
	Telepon    string
	Email      string
	Kategori   string
	Subjek     string
	Pesan      string
	Status     string
	Catatan    string
	IPAddress  string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type ContactRepository interface {
	Create(ctx context.Context, m *ContactMessage) error
	ByTicket(ctx context.Context, code string) (*ContactMessage, error)
	// CountRecentFromIP backs submission rate limiting.
	CountRecentFromIP(ctx context.Context, ip string, since time.Time) (int64, error)
}
