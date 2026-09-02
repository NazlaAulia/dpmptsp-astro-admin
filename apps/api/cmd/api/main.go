// Command api is the DPMPTSP backend.
//
// It is reachable only from the internal docker network (SPEC.md §8): the Astro
// apps call it server-side, and nothing else does. It owns all database access
// and, once the domain handlers land, all business logic and authorization
// (CLAUDE.md rules 2 and 3).
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"dpmptsp/api/internal/config"
	"dpmptsp/api/internal/http/routes"
	"dpmptsp/api/internal/infrastructure/cache"
	"dpmptsp/api/internal/infrastructure/database"
	"dpmptsp/api/internal/infrastructure/storage"
)

func main() {
	if err := run(); err != nil {
		slog.Error("startup failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(log)

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Bound the whole startup: a database that never answers should fail the
	// container, not hang it in a state where orchestration thinks it is fine.
	startupCtx, cancelStartup := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelStartup()

	db, dia, err := database.Open(startupCtx, cfg, log)
	if err != nil {
		return err
	}
	defer database.Close(db)
	log.Info("database connected", "engine", dia.Name(), "target", cfg.DB.Redacted())

	// Redis is optional at boot. The cache is an optimisation (SPEC.md §6), and
	// refusing to serve because a cache is down would turn a performance
	// dependency into an availability one.
	var rdb *cache.Client
	if rdb, err = cache.Open(startupCtx, cfg.RedisURL); err != nil {
		log.Warn("redis unavailable, continuing without cache", "error", err)
		rdb = nil
	} else {
		defer rdb.Close()
		log.Info("redis connected")
	}

	// Storage disks. Configured like Laravel's filesystem: several disks
	// defined, one selected by FILESYSTEM_DISK. Built at startup so a missing
	// bucket or an unwritable directory stops the container instead of
	// surfacing on the first upload.
	files, err := storage.FromEnv(startupCtx)
	if err != nil {
		return err
	}
	log.Info("storage ready", "default", files.Default().Name(), "disks", files.Names())

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           routes.New(cfg, log, db, dia, rdb, files),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info("listening", "addr", cfg.Addr, "env", cfg.Env)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return err
	case sig := <-stop:
		log.Info("shutting down", "signal", sig.String())
	}

	// Drain in-flight requests rather than cutting them off mid-response.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}
	log.Info("stopped cleanly")
	return nil
}
