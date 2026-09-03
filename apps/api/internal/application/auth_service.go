package application

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"dpmptsp/api/internal/domain"
	"dpmptsp/api/internal/security"
)

// SessionTTL bounds how long a signed-in session lasts.
const SessionTTL = 8 * time.Hour

// ErrInvalidCredentials is returned for every authentication failure.
//
// One error for all cases: distinguishing "no such user" from "wrong password"
// turns the login form into a username oracle.
var ErrInvalidCredentials = errors.New("invalid credentials")

type AuthService struct {
	users    domain.UserRepository
	sessions domain.SessionStore
}

func NewAuthService(u domain.UserRepository, s domain.SessionStore) *AuthService {
	return &AuthService{users: u, sessions: s}
}

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

// Login authenticates and opens a session, returning its id.
func (s *AuthService) Login(ctx context.Context, username, password string) (*domain.Session, error) {
	u, err := s.Authenticate(ctx, username, password)
	if err != nil {
		return nil, err
	}

	id, err := newSessionID()
	if err != nil {
		return nil, err
	}
	now := time.Now()
	sess := domain.Session{
		ID: id, UserID: u.ID, Username: u.Username, Role: u.Role,
		CreatedAt: now, ExpiresAt: now.Add(SessionTTL),
	}
	if err := s.sessions.Create(ctx, sess); err != nil {
		return nil, fmt.Errorf("open session: %w", err)
	}
	return &sess, nil
}

// Session resolves a session id.
func (s *AuthService) Session(ctx context.Context, id string) (*domain.Session, error) {
	if strings.TrimSpace(id) == "" {
		return nil, domain.ErrNotFound
	}
	return s.sessions.Get(ctx, id)
}

// Logout revokes one session.
func (s *AuthService) Logout(ctx context.Context, id string) error {
	return s.sessions.Delete(ctx, id)
}

func newSessionID() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate session id: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
