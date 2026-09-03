package middleware

import (
	"context"
	"net/http"

	"dpmptsp/api/internal/domain"
)

const sessionKey ctxKey = "session"

// SessionResolver looks up a session id.
type SessionResolver interface {
	Session(ctx context.Context, id string) (*domain.Session, error)
}

// Session resolves the X-Session-Id header onto the request context.
//
// It never rejects: a request without a valid session simply carries none.
// RequireRole decides what that means for a given route.
func Session(resolver SessionResolver) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := r.Header.Get("X-Session-Id")
			if id != "" {
				if s, err := resolver.Session(r.Context(), id); err == nil {
					r = r.WithContext(context.WithValue(r.Context(), sessionKey, s))
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// SessionFrom returns the resolved session, if any.
func SessionFrom(ctx context.Context) *domain.Session {
	if s, ok := ctx.Value(sessionKey).(*domain.Session); ok {
		return s
	}
	return nil
}

// RequireRole refuses a request that has no session, or whose session does not
// hold one of the given roles.
//
// This is where authorization lives. The service key only proves the caller is
// one of our own applications; it says nothing about which person is acting,
// which is why it cannot be the only check on a mutating route.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			s := SessionFrom(r.Context())
			if s == nil {
				writeJSONError(w, http.StatusUnauthorized, "authentication required")
				return
			}
			if len(allowed) > 0 && !allowed[s.Role] {
				writeJSONError(w, http.StatusForbidden, "insufficient role")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func writeJSONError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write([]byte(`{"error":"` + msg + `"}`))
}
