package config

import (
	"strings"
	"testing"
)

const validKey = "0123456789abcdef0123456789abcdef"

// Database settings are discrete, as in Laravel — DB_CONNECTION, DB_HOST,
// DB_PORT, DB_DATABASE, DB_USERNAME, DB_PASSWORD — and the driver-specific DSN
// is built from them. Asking an operator for a raw DSN would leak the engine
// switch into the environment file, which is the thing it exists to hide.

func setValid(t *testing.T) {
	t.Helper()
	t.Setenv("API_SERVICE_KEY", validKey)
	t.Setenv("DB_DATABASE", "ladpm")
	t.Setenv("DB_USERNAME", "root")
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("DB_PORT", "")
	t.Setenv("DB_HOST", "")
	t.Setenv("DB_CONNECTION", "")
	t.Setenv("DB_ENGINE", "")
}

func TestLoadReportsEveryProblemAtOnce(t *testing.T) {
	t.Setenv("API_SERVICE_KEY", "")
	t.Setenv("DB_DATABASE", "")
	t.Setenv("DB_USERNAME", "")
	t.Setenv("DATABASE_URL", "")

	_, err := Load()
	if err == nil {
		t.Fatal("expected configuration to be rejected")
	}
	// One restart per mistake is miserable when a stack has just been stood up.
	for _, want := range []string{"DB_DATABASE", "DB_USERNAME", "API_SERVICE_KEY"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not mention %s", err, want)
		}
	}
}

func TestPortDefaultsToTheEnginesPort(t *testing.T) {
	setValid(t)
	t.Setenv("DB_CONNECTION", "postgres")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DB.Port != 5432 {
		t.Errorf("postgres port = %d, want 5432", cfg.DB.Port)
	}

	t.Setenv("DB_CONNECTION", "mysql")
	cfg, err = Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DB.Port != 3306 {
		t.Errorf("mysql port = %d, want 3306", cfg.DB.Port)
	}
}

func TestDBEngineIsAcceptedAsAnAlias(t *testing.T) {
	setValid(t)
	t.Setenv("DB_ENGINE", "mysql")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Engine() != MySQL {
		t.Errorf("engine = %q, want mysql", cfg.Engine())
	}
}

func TestRejectsUnknownConnection(t *testing.T) {
	setValid(t)
	t.Setenv("DB_CONNECTION", "sqlite")
	if _, err := Load(); err == nil || !strings.Contains(err.Error(), "DB_CONNECTION") {
		t.Fatalf("expected DB_CONNECTION to be rejected, got %v", err)
	}
}

func TestRejectsShortServiceKey(t *testing.T) {
	setValid(t)
	t.Setenv("API_SERVICE_KEY", "tooshort")
	if _, err := Load(); err == nil || !strings.Contains(err.Error(), "32 characters") {
		t.Fatalf("expected a length complaint, got %v", err)
	}
}

func TestPostgresDSN(t *testing.T) {
	d := Database{
		Connection: Postgres, Host: "database", Port: 5432,
		Name: "ladpm", Username: "root", Password: "secret",
	}
	got, err := d.DSN()
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"postgres://", "root:secret@database:5432/ladpm", "sslmode=disable"} {
		if !strings.Contains(got, want) {
			t.Errorf("DSN %q missing %q", got, want)
		}
	}
}

func TestMySQLDSNIsNotAURL(t *testing.T) {
	d := Database{
		Connection: MySQL, Host: "database", Port: 3306,
		Name: "ladpm", Username: "root", Password: "secret",
	}
	got, err := d.DSN()
	if err != nil {
		t.Fatal(err)
	}
	if strings.HasPrefix(got, "mysql://") {
		t.Errorf("go-sql-driver does not accept a URL scheme: %q", got)
	}
	for _, want := range []string{"root:secret@tcp(database:3306)/ladpm", "parseTime=true", "charset=utf8mb4"} {
		if !strings.Contains(got, want) {
			t.Errorf("DSN %q missing %q", got, want)
		}
	}
}

// A password with @ or / breaks naive string building, and passwords like this
// are exactly what a generated secret looks like.
func TestPostgresDSNEscapesAwkwardPasswords(t *testing.T) {
	d := Database{
		Connection: Postgres, Host: "db", Port: 5432,
		Name: "app", Username: "user", Password: "p@ss/w0rd?x",
	}
	got, err := d.DSN()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(got, "p@ss/w0rd?x") {
		t.Errorf("password was not escaped: %q", got)
	}
}

func TestDATABASE_URLOverridesEverything(t *testing.T) {
	d := Database{
		Connection: Postgres, Host: "ignored", Port: 1,
		Name: "ignored", Username: "ignored",
		URL: "postgres://someone@managed.example/db",
	}
	got, err := d.DSN()
	if err != nil {
		t.Fatal(err)
	}
	if got != "postgres://someone@managed.example/db" {
		t.Errorf("got %q", got)
	}
}

func TestRedactedNeverContainsThePassword(t *testing.T) {
	d := Database{
		Connection: MySQL, Host: "database", Port: 3306,
		Name: "ladpm", Username: "root", Password: "sup3rs3cret",
	}
	if strings.Contains(d.Redacted(), "sup3rs3cret") {
		t.Errorf("password leaked into logs: %q", d.Redacted())
	}
}
