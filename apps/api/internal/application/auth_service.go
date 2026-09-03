package application

import (
	"context"
	"errors"
	"strings"

	"dpmptsp/api/internal/domain"
	"dpmptsp/api/internal/security"
)

// ErrInvalidCredentials is returned for every authentication failure.
//
// One error for all cases: distinguishing "no such user" from "wrong password"
// turns the login form into a username oracle.
var ErrInvalidCredentials = errors.New("invalid credentials")

type AuthService struct {
	users domain.UserRepository
}

func NewAuthService(u domain.UserRepository) *AuthService { return &AuthService{users: u} }

// Authenticate verifies a username and password.
func (s *AuthService) Authenticate(ctx context.Context, username, password string) (*domain.User, error) {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return nil, ErrInvalidCredentials
	}

	u, err := s.users.ByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			// Hash anyway so a missing user takes the same time as a wrong
			// password; returning early is a timing oracle for valid usernames.
			_, _ = security.Hash(password)
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if !u.IsActive {
		return nil, ErrInvalidCredentials
	}
	if err := security.Verify(password, u.PasswordHash); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Upgrade the stored hash if the cost or algorithm has since changed.
	if security.NeedsRehash(u.PasswordHash) {
		if newHash, err := security.Hash(password); err == nil {
			_ = s.users.UpdatePasswordHash(ctx, u.ID, newHash)
		}
	}
	_ = s.users.TouchLogin(ctx, u.ID)

	u.PasswordHash = ""
	return u, nil
}
