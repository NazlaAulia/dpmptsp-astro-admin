package config

import (
	"fmt"
	"net/url"
	"strings"
)

// Database holds the discrete connection settings:
//
//	DB_CONNECTION DB_HOST DB_PORT DB_DATABASE DB_USERNAME DB_PASSWORD
//
// The driver-specific DSN is built from them by DSN, since pgx expects a URL
// and go-sql-driver expects user:pass@tcp(host:port)/db.
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

	// URL, when set, is used verbatim and the fields above are ignored.
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
		// go-sql-driver does not accept a URL scheme.
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
