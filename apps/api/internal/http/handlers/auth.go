package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"dpmptsp/api/internal/application"
	"dpmptsp/api/internal/domain"
	"dpmptsp/api/internal/http/middleware"
	"dpmptsp/api/internal/http/render"
)

type Auth struct {
	Service *application.AuthService
	Log     *slog.Logger
}

type loginInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type sessionDTO struct {
	SessionID string `json:"session_id,omitempty"`
	UserID    int64  `json:"user_id"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	ExpiresAt string `json:"expires_at"`
}

// Login verifies credentials and opens a session.
//
// The session id is returned once, here. Astro stores it in its cookie and
// sends it back as X-Session-Id; the API resolves it against Redis on every
// request, which is what makes logout able to revoke it.
func (h *Auth) Login(w http.ResponseWriter, r *http.Request) {
	var in loginInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		render.BadRequest(w, "body must be valid JSON")
		return
	}

	sess, err := h.Service.Login(r.Context(), in.Username, in.Password)
	if err != nil {
		if errors.Is(err, application.ErrInvalidCredentials) {
			render.JSON(w, http.StatusUnauthorized, map[string]string{
				"error": "Username atau password salah.",
			})
			return
		}
		render.Error(w, h.Log, middleware.RequestIDFrom(r.Context()), err)
		return
	}

	render.OK(w, sessionDTO{
		SessionID: sess.ID, UserID: sess.UserID, Username: sess.Username,
		Role: sess.Role, ExpiresAt: sess.ExpiresAt.UTC().Format("2006-01-02T15:04:05Z"),
	})
}

// Current reports who the caller is, for the session resolved by middleware.
func (h *Auth) Current(w http.ResponseWriter, r *http.Request) {
	s := middleware.SessionFrom(r.Context())
	if s == nil {
		render.JSON(w, http.StatusUnauthorized, map[string]string{"error": "no session"})
		return
	}
	render.OK(w, sessionDTO{
		UserID: s.UserID, Username: s.Username, Role: s.Role,
		ExpiresAt: s.ExpiresAt.UTC().Format("2006-01-02T15:04:05Z"),
	})
}

// Logout revokes the session named by X-Session-Id.
func (h *Auth) Logout(w http.ResponseWriter, r *http.Request) {
	id := r.Header.Get("X-Session-Id")
	if id == "" {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if err := h.Service.Logout(r.Context(), id); err != nil && !errors.Is(err, domain.ErrNotFound) {
		render.Error(w, h.Log, middleware.RequestIDFrom(r.Context()), err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
