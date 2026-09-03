// Package config reads and validates every setting the API needs, once, at
// startup. Anything missing or malformed fails here rather than surfacing as a
// confusing error on the first request that happens to need it.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Engine is the SQL dialect the API talks to. Postgres is the deployment
// target; MySQL is supported and verified.
type Engine string

const (
	Postgres Engine = "postgres"
	MySQL    Engine = "mysql"
)

type Config struct {
	Addr            string
	DB              Database
	RedisURL        string
	ServiceKey      string
	ShutdownTimeout time.Duration
	Env             string
}

// Engine reports the configured SQL dialect.
func (c *Config) Engine() Engine { return c.DB.Connection }

// Load reads configuration from the environment and reports every problem at
// once rather than failing on the first.
func Load() (*Config, error) {
	var problems []string

	// DB_ENGINE is accepted as an alias for DB_CONNECTION.
	connection := envOr("DB_CONNECTION", envOr("DB_ENGINE", string(Postgres)))

	cfg := &Config{
		Addr: envOr("API_ADDR", ":8080"),
		DB: Database{
			Connection: Engine(strings.ToLower(connection)),
			Host:       envOr("DB_HOST", "database"),
			Port:       0, // filled in below, per driver
			Name:       os.Getenv("DB_DATABASE"),
			Username:   os.Getenv("DB_USERNAME"),
			Password:   os.Getenv("DB_PASSWORD"),
			SSLMode:    os.Getenv("DB_SSLMODE"),
			Charset:    os.Getenv("DB_CHARSET"),
			URL:        os.Getenv("DATABASE_URL"),
		},
		RedisURL:        envOr("REDIS_URL", "redis://redis:6379/0"),
		ServiceKey:      os.Getenv("API_SERVICE_KEY"),
		ShutdownTimeout: 15 * time.Second,
		Env:             envOr("APP_ENV", "development"),
	}

	switch cfg.DB.Connection {
	case Postgres, MySQL:
	default:
		problems = append(problems, fmt.Sprintf(
			"DB_CONNECTION must be %q or %q, got %q", Postgres, MySQL, cfg.DB.Connection))
	}

	// Default the port to the engine's standard port.
	defaultPort := 5432
	if cfg.DB.Connection == MySQL {
		defaultPort = 3306
	}
	if v := os.Getenv("DB_PORT"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 || n > 65535 {
			problems = append(problems, "DB_PORT must be a valid port number")
		} else {
			cfg.DB.Port = n
		}
	} else {
		cfg.DB.Port = defaultPort
	}

	// DATABASE_URL replaces the discrete settings when set.
	if cfg.DB.URL == "" {
		if cfg.DB.Name == "" {
			problems = append(problems, "DB_DATABASE is required (or set DATABASE_URL)")
		}
		if cfg.DB.Username == "" {
			problems = append(problems, "DB_USERNAME is required (or set DATABASE_URL)")
		}
	}

	// The API is only ever called by the Astro apps over the internal docker
	// network, and that call is authenticated with a shared key
	//. Refusing to start without one keeps "temporarily unset"
	// from silently becoming "unauthenticated in production".
	if cfg.ServiceKey == "" {
		problems = append(problems, "API_SERVICE_KEY is required")
	} else if len(cfg.ServiceKey) < 32 {
		problems = append(problems, "API_SERVICE_KEY must be at least 32 characters")
	}

	if v := os.Getenv("SHUTDOWN_TIMEOUT_SECONDS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			problems = append(problems, "SHUTDOWN_TIMEOUT_SECONDS must be a positive integer")
		} else {
			cfg.ShutdownTimeout = time.Duration(n) * time.Second
		}
	}

	if len(problems) > 0 {
		return nil, fmt.Errorf("invalid configuration:\n  - %s", strings.Join(problems, "\n  - "))
	}
	return cfg, nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
