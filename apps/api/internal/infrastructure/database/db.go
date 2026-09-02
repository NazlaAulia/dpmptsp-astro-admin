// Package database opens the GORM connection and pairs it with the dialect
// covering what GORM does not abstract away.
package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"dpmptsp/api/internal/config"
	"dpmptsp/api/internal/infrastructure/database/dialect"
)

// Open connects and verifies the connection before returning it, so a bad DSN
// is a startup failure rather than a runtime surprise.
//
// GORM carries most of the engine difference: placeholders, RETURNING vs
// LastInsertId, and upsert clauses are all rendered per driver. What it does
// not cover stays in the dialect package.
func Open(ctx context.Context, cfg *config.Config, log *slog.Logger) (*gorm.DB, dialect.Dialect, error) {
	dsn, err := cfg.DB.DSN()
	if err != nil {
		return nil, nil, err
	}

	var (
		opener gorm.Dialector
		d      dialect.Dialect
	)
	switch cfg.Engine() {
	case config.Postgres:
		opener, d = postgres.Open(dsn), dialect.Postgres{}
	case config.MySQL:
		opener, d = mysql.Open(dsn), dialect.MySQL{}
	default:
		return nil, nil, fmt.Errorf("unsupported database engine %q", cfg.Engine())
	}

	level := gormlogger.Warn
	if cfg.Env == "development" {
		level = gormlogger.Info
	}

	db, err := gorm.Open(opener, &gorm.Config{
		Logger: gormlogger.Default.LogMode(level),
		// The schema is owned by golang-migrate (SPEC.md §4: migrations are the
		// source of truth). GORM must never alter it, so automigration is not
		// used anywhere and naming is left exactly as the migrations declare.
		DisableAutomaticPing:                     true,
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("open %s: %w", d.Name(), err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, fmt.Errorf("underlying pool: %w", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)

	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(pingCtx); err != nil {
		sqlDB.Close()
		return nil, nil, fmt.Errorf("ping %s: %w", d.Name(), err)
	}

	return db, d, nil
}

// Close shuts the underlying pool down. GORM has no Close of its own.
func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Pool exposes the standard-library handle, for the health check.
func Pool(db *gorm.DB) (*sql.DB, error) { return db.DB() }
