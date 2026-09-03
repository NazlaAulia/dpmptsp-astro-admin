package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"dpmptsp/api/internal/application"
	"dpmptsp/api/internal/domain"
	"dpmptsp/api/internal/http/middleware"
	"dpmptsp/api/internal/http/render"
)

type Contact struct {
	Service *application.ContactService
	Log     *slog.Logger
}

type contactInput struct {
	Nama     string `json:"nama"`
	Telepon  string `json:"telepon"`
	Email    string `json:"email"`
	Kategori string `json:"kategori"`
	Subjek   string `json:"subjek"`
	Pesan    string `json:"pesan"`
	// Honeypot. Real users never see this field, so anything in it is a bot.
	Website string `json:"website"`
}

type contactCreatedDTO struct {
	Tiket     string `json:"tiket"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

// trackDTO is what the tracking page shows back to a reporter.
//
// The ticket code is the only credential here, so the contact details are held
// back: email is never returned at all and the phone number is masked. The code
// travels in a query string and can end up in a log or a shared screenshot,
// which is not a good enough reason to hand out a working phone number.
type trackDTO struct {
	Tiket     string `json:"tiket"`
	Status    string `json:"status"`
	Nama      string `json:"nama,omitempty"`
	Telepon   string `json:"telepon,omitempty"`
	Kategori  string `json:"kategori,omitempty"`
	Subjek    string `json:"subjek,omitempty"`
	Pesan     string `json:"pesan,omitempty"`
	Catatan   string `json:"catatan,omitempty"`
	Tanggal   string `json:"tanggal"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// maskPhone keeps enough for the reporter to recognise their own number and
// too little for anyone else to dial it.
func maskPhone(s string) string {
	if len(s) <= 4 {
		return strings.Repeat("*", len(s))
	}
	return s[:2] + strings.Repeat("*", len(s)-4) + s[len(s)-2:]
}

func (h *Contact) Create(w http.ResponseWriter, r *http.Request) {
	var in contactInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		render.BadRequest(w, "body must be valid JSON")
		return
	}

	// Answer a filled honeypot exactly like a success, so a bot cannot tell it
	// was caught and retry differently.
	if strings.TrimSpace(in.Website) != "" {
		render.Created(w, contactCreatedDTO{Tiket: "DPM-00000000-XXXXXXXX", Status: "baru"})
		return
	}

	msg := &domain.ContactMessage{
		Nama: in.Nama, Telepon: in.Telepon, Email: in.Email,
		Kategori: in.Kategori, Subjek: in.Subjek, Pesan: in.Pesan,
		IPAddress: clientIP(r),
	}

	if err := h.Service.Submit(r.Context(), msg); err != nil {
		if errors.Is(err, application.ErrRateLimited) {
			render.JSON(w, http.StatusTooManyRequests, map[string]string{
				"error": "Terlalu banyak pengiriman. Coba lagi nanti.",
			})
			return
		}
		render.Error(w, h.Log, middleware.RequestIDFrom(r.Context()), err)
		return
	}

	render.Created(w, contactCreatedDTO{
		Tiket:     msg.TicketCode,
		Status:    msg.Status,
		CreatedAt: msg.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	})
}

func (h *Contact) Track(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("tiket")

	msg, err := h.Service.Track(r.Context(), code)
	if err != nil {
		// A missing ticket and a malformed one answer identically, so the
		// endpoint cannot be used to probe which codes exist.
		if errors.Is(err, domain.ErrNotFound) {
			render.JSON(w, http.StatusNotFound, map[string]string{
				"error": "Nomor tiket tidak ditemukan.",
			})
			return
		}
		render.Error(w, h.Log, middleware.RequestIDFrom(r.Context()), err)
		return
	}

	created := msg.CreatedAt.UTC().Format("2006-01-02T15:04:05Z")
	render.OK(w, trackDTO{
		Tiket: msg.TicketCode, Status: msg.Status,
		Nama: msg.Nama, Telepon: maskPhone(msg.Telepon),
		Kategori: msg.Kategori, Subjek: msg.Subjek, Pesan: msg.Pesan,
		Catatan: msg.Catatan,
		Tanggal: created, CreatedAt: created,
		UpdatedAt: msg.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	})
}

// clientIP prefers the address the gateway forwarded.
func clientIP(r *http.Request) string {
	if v := r.Header.Get("X-Forwarded-For"); v != "" {
		if i := strings.IndexByte(v, ','); i > 0 {
			return strings.TrimSpace(v[:i])
		}
		return strings.TrimSpace(v)
	}
	return r.RemoteAddr
}
