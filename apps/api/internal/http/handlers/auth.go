package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"dpmptsp/api/internal/application"
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

type userDTO struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email,omitempty"`
	Role     string `json:"role"`
}

// Login verifies credentials. Astro holds the session cookie; this only says
// whether the credentials are valid and who they belong to.
func (h *Auth) Login(w http.ResponseWriter, r *http.Request) {
	var in loginInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		render.BadRequest(w, "body must be valid JSON")
		return
	}

	u, err := h.Service.Authenticate(r.Context(), in.Username, in.Password)
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

	render.OK(w, userDTO{ID: u.ID, Username: u.Username, Email: u.Email, Role: u.Role})
}
