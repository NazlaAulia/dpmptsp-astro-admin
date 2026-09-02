// Package handlers holds the HTTP handlers.
package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"gorm.io/gorm"

	"dpmptsp/api/internal/infrastructure/cache"
	"dpmptsp/api/internal/infrastructure/storage"
)

type Health struct {
	DB     *gorm.DB
	Redis  *cache.Client
	Engine string
	Files  *storage.Manager
}

type healthResponse struct {
	Status  string            `json:"status"`
	Engine  string            `json:"engine"`
	Storage string            `json:"storage"`
	Checks  map[string]string `json:"checks"`
}

// Live answers "is the process up?" and touches no dependency. Container
// liveness probes must not restart the API because the database blinked.
func (h *Health) Live(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Ready answers "can this instance serve traffic?" and therefore does check the
// dependencies.
func (h *Health) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	resp := healthResponse{
		Status: "ok",
		Engine: h.Engine,
		Checks: map[string]string{},
	}
	if h.Files != nil {
		resp.Storage = h.Files.Default().Name()
	}
	code := http.StatusOK

	if err := pingDB(ctx, h.DB); err != nil {
		resp.Checks["database"] = "error: " + err.Error()
		resp.Status = "degraded"
		code = http.StatusServiceUnavailable
	} else {
		resp.Checks["database"] = "ok"
	}

	if h.Redis == nil {
		resp.Checks["redis"] = "disabled"
	} else if err := h.Redis.Ping(ctx); err != nil {
		resp.Checks["redis"] = "error: " + err.Error()
		resp.Status = "degraded"
		code = http.StatusServiceUnavailable
	} else {
		resp.Checks["redis"] = "ok"
	}

	writeJSON(w, code, resp)
}

// pingDB reaches through GORM to the standard-library pool, which is what
// actually knows whether the connection is alive.
func pingDB(ctx context.Context, db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

func writeJSON(w http.ResponseWriter, code int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(body)
}
