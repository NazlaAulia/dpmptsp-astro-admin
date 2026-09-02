package config

import (
	"fmt"
	"net/url"
	"strings"
)

// Database is configured with discrete variables, the way Laravel does it:
//
//	DB_CONNECTION=mysql
//	DB_HOST=database
//	DB_PORT=3306
//	DB_DATABASE=ladpm
//	DB_USERNAME=root
//	DB_PASSWORD=secret
//
// The driver-specific DSN is then built here. That matters because the two
// drivers disagree about what a DSN even looks like — pgx wants a URL, and
// go-sql-driver/mysql wants `user:pass@tcp(host:port)/db` — so asking an
// operator to supply one directly leaks the engine switch straight into the
// environment file, which is exactly what it exists to hide.
type Database struct {
	Connection Engine
	Host       string
	Port       int
	Name       string
	Username   string
	Password   string

	// SSLMode is Postgres only: disable, require, verify-ca, verify-full.
	SSLMode string
	// Charset is MySQL only.
	Charset string

	// URL, when set, is used verbatim and every field above is ignored.
	// Laravel offers the same escape hatch, and it is the practical way to
	// accept a connection string from a managed provider.
	URL string
}

// DSN renders the connection string for the configured driver.
func (d Database) DSN() (string, error) {
	if d.URL != "" {
		return d.URL, nil
	}

	switch d.Connection {
	case Postgres:
		u := url.URL{
			Scheme: "postgres",
			User:   url.UserPassword(d.Username, d.Password),
			Host:   fmt.Sprintf("%s:%d", d.Host, d.Port),
			Path:   "/" + d.Name,
		}
		q := url.Values{}
		q.Set("sslmode", orDefault(d.SSLMode, "disable"))
		u.RawQuery = q.Encode()
		return u.String(), nil

	case MySQL:
		// Not a URL. Credentials are escaped by the driver's own rules, and a
		// password containing '@' or '/' would otherwise break parsing.
		q := url.Values{}
		q.Set("parseTime", "true")
		q.Set("charset", orDefault(d.Charset, "utf8mb4"))
		q.Set("loc", "UTC")
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s",
			d.Username, d.Password, d.Host, d.Port, d.Name, q.Encode()), nil

	default:
		return "", fmt.Errorf("unsupported database connection %q", d.Connection)
	}
}

// Redacted renders the DSN with the password replaced, for logs.
func (d Database) Redacted() string {
	if d.URL != "" {
		return "(DATABASE_URL)"
	}
	return fmt.Sprintf("%s://%s@%s:%d/%s", d.Connection, d.Username, d.Host, d.Port, d.Name)
}

func orDefault(v, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}
