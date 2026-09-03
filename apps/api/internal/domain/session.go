package domain

import (
	"context"
	"time"
)

// Session is a signed-in administrator, held server-side.
//
// Server-side storage is what makes logout and revocation possible: a
// self-contained token stays valid until it expires no matter what the server
// does with it.
type Session struct {
	ID        string
	UserID    int64
	Username  string
	Role      string
	CreatedAt time.Time
	ExpiresAt time.Time
}

type SessionStore interface {
	Create(ctx context.Context, s Session) error
	Get(ctx context.Context, id string) (*Session, error)
	Delete(ctx context.Context, id string) error
	// DeleteForUser revokes every session a user holds, for "sign out
	// everywhere" and for locking out a compromised account.
	DeleteForUser(ctx context.Context, userID int64) error
}
