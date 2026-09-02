// Package render is the single place responses are written, so every endpoint
// answers in the same shape and no handler invents its own error format.
package render

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"dpmptsp/api/internal/domain"
)

type Envelope struct {
	Data any   `json:"data,omitempty"`
	Meta *Meta `json:"meta,omitempty"`
}

type Meta struct {
	Total   int64 `json:"total"`
	Page    int   `json:"page"`
	PerPage int   `json:"per_page"`
	Pages   int   `json:"pages"`
}

type errorBody struct {
	Error string `json:"error"`
}

func JSON(w http.ResponseWriter, code int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	if body != nil {
		_ = json.NewEncoder(w).Encode(body)
	}
}

func OK(w http.ResponseWriter, data any)      { JSON(w, http.StatusOK, Envelope{Data: data}) }
func Created(w http.ResponseWriter, data any) { JSON(w, http.StatusCreated, Envelope{Data: data}) }

func Paged(w http.ResponseWriter, data any, total int64, page, perPage int) {
	pages := 0
	if perPage > 0 {
		pages = int((total + int64(perPage) - 1) / int64(perPage))
	}
	JSON(w, http.StatusOK, Envelope{
		Data: data,
		Meta: &Meta{Total: total, Page: page, PerPage: perPage, Pages: pages},
	})
}

// Error maps a domain error onto a status code and a message safe to send.
//
// The internal error text is logged, never returned: a database error string
// can name tables, columns and constraints, and there is no reason to hand that
// to a caller.
func Error(w http.ResponseWriter, log *slog.Logger, requestID string, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		JSON(w, http.StatusNotFound, errorBody{"not found"})
	case errors.Is(err, domain.ErrInvalid):
		// Validation messages are written by us and are safe to show.
		JSON(w, http.StatusUnprocessableEntity, errorBody{err.Error()})
	case errors.Is(err, domain.ErrConflict):
		JSON(w, http.StatusConflict, errorBody{err.Error()})
	default:
		log.Error("request failed", "error", err, "request_id", requestID)
		JSON(w, http.StatusInternalServerError, errorBody{"internal error"})
	}
}

func BadRequest(w http.ResponseWriter, msg string) {
	JSON(w, http.StatusBadRequest, errorBody{msg})
}
