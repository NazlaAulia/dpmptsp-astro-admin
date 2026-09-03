package domain

import (
	"context"
	"time"
)

// User is an administrator of the CMS.
type User struct {
	ID          int64
	Username    string
	Email       string
	Role        string
	IsActive    bool
	LastLoginAt *time.Time

	// PasswordHash never leaves the domain or repository layers.
	PasswordHash string
}

type UserRepository interface {
	ByUsername(ctx context.Context, username string) (*User, error)
	TouchLogin(ctx context.Context, id int64) error
	UpdatePasswordHash(ctx context.Context, id int64, hash string) error
}
