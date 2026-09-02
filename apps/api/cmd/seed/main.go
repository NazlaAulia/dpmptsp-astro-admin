// Command seed populates reference data, in the shape `php artisan db:seed`
// takes: seed files are declarative, they run in a defined order, and running
// them twice is safe.
//
//	seed              run every seeder
//	seed -only=users  run one, by filename fragment
//	seed -list        show what would run
//	seed -fresh       delete the rows a seeder owns, then re-insert them
//
// Seeds are YAML rather than SQL on purpose. SQL would have to be written twice,
// once per dialect, and would drift; a declarative row set is rendered into the
// right INSERT for whichever engine is configured.
//
// Reference and configuration data belongs here. The 553 legacy articles do NOT
// — that is a one-shot import of production content, not a seeder, and it is
// kept separate per SPEC.md §4.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
	"gorm.io/gorm"

	"dpmptsp/api/internal/config"
	"dpmptsp/api/internal/infrastructure/database"
	"dpmptsp/api/internal/infrastructure/database/dialect"
	"dpmptsp/api/internal/security"
)

// Seeder is one YAML file: a table, the columns that identify a row, and the
// rows it owns.
type Seeder struct {
	Table string `yaml:"table"`
	// Key names the columns that make a row unique, so a re-run can tell an
	// existing row from a new one. Declared explicitly rather than assumed to
	// be the primary key: some seeders identify rows by a natural key such as
	// username.
	Key  []string         `yaml:"key"`
	Rows []map[string]any `yaml:"rows"`

	name string // source filename, for output
}

func main() {
	var (
		dir   = flag.String("dir", "seeds", "directory holding the seed files")
		only  = flag.String("only", "", "run only seeders whose filename contains this")
		list  = flag.Bool("list", false, "list the seeders that would run, and do nothing")
		fresh = flag.Bool("fresh", false, "delete the rows each seeder owns before inserting")
	)
	flag.Parse()

	if err := run(*dir, *only, *list, *fresh); err != nil {
		fmt.Fprintf(os.Stderr, "seed failed: %v\n", err)
		os.Exit(1)
	}
}

func run(dir, only string, list, fresh bool) error {
	seeders, err := load(dir, only)
	if err != nil {
		return err
	}
	if len(seeders) == 0 {
		return fmt.Errorf("no seed files matched in %s", dir)
	}

	if list {
		for _, s := range seeders {
			fmt.Printf("  %-32s %-28s %d rows\n", s.name, s.Table, len(s.Rows))
		}
		return nil
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	db, dia, err := database.Open(ctx, cfg, slog.Default())
	if err != nil {
		return err
	}
	defer database.Close(db)

	fmt.Printf("seeding %s (%s)\n", cfg.DB.Redacted(), dia.Name())

	for _, s := range seeders {
		inserted, skipped, err := s.apply(ctx, db, dia, fresh)
		if err != nil {
			return fmt.Errorf("%s: %w", s.name, err)
		}
		fmt.Printf("  %-32s %d inserted, %d already present\n", s.name, inserted, skipped)
	}
	fmt.Println("done")
	return nil
}

func load(dir, only string) ([]Seeder, error) {
	entries, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		return nil, err
	}
	// Filename order is run order, which is why the files are numbered: a table
	// with a foreign key must be seeded after the table it points at.
	sort.Strings(entries)

	var out []Seeder
	for _, path := range entries {
		base := filepath.Base(path)
		if only != "" && !strings.Contains(base, only) {
			continue
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		var s Seeder
		if err := yaml.Unmarshal(raw, &s); err != nil {
			return nil, fmt.Errorf("%s: %w", base, err)
		}
		if s.Table == "" {
			return nil, fmt.Errorf("%s: no table declared", base)
		}
		if len(s.Key) == 0 {
			return nil, fmt.Errorf("%s: no key declared — a seeder must say which "+
				"columns identify a row, or re-running it would duplicate data", base)
		}
		s.name = base
		out = append(out, s)
	}
	return out, nil
}

// apply inserts the rows, skipping any that are already there.
//
// Idempotence is the point. Laravel's seeders are not idempotent by default and
// re-running them duplicates data; here a second run is a no-op, so `make seed`
// is safe to repeat and safe to put in a setup script.
func (s Seeder) apply(ctx context.Context, db *gorm.DB, _ dialect.Dialect, fresh bool) (int, int, error) {
	var inserted, skipped int

	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if fresh {
			if err := tx.Exec("DELETE FROM " + tx.Statement.Quote(s.Table)).Error; err != nil {
				return fmt.Errorf("clearing %s: %w", s.Table, err)
			}
		}

		for i, row := range s.Rows {
			values, err := resolve(row)
			if err != nil {
				return fmt.Errorf("row %d: %w", i+1, err)
			}

			// Look the row up by its declared key, then insert only if absent.
			//
			// This replaced a conflict clause. GORM's OnConflict{DoNothing} is
			// not portable here: on MySQL it renders "ON DUPLICATE KEY UPDATE"
			// with an empty assignment list, which is a syntax error. An
			// existence check costs one extra SELECT per row and works
			// identically on both engines.
			where := tx.Table(s.Table)
			for _, k := range s.Key {
				v, ok := values[k]
				if !ok {
					return fmt.Errorf("row %d: key column %q is not present in the row", i+1, k)
				}
				where = where.Where(tx.Statement.Quote(k)+" = ?", v)
			}

			var count int64
			if err := where.Count(&count).Error; err != nil {
				return fmt.Errorf("row %d: %w", i+1, err)
			}
			if count > 0 {
				skipped++
				continue
			}

			if err := tx.Table(s.Table).Create(values).Error; err != nil {
				return fmt.Errorf("row %d: %w", i+1, err)
			}
			inserted++
		}
		return nil
	})

	return inserted, skipped, err
}

// resolve expands ${VAR} references so a secret is read from the environment at
// seed time rather than being committed to a YAML file.
func resolve(row map[string]any) (map[string]any, error) {
	out := make(map[string]any, len(row))
	for c, v := range row {
		s, ok := v.(string)
		if !ok || !strings.HasPrefix(s, "${") || !strings.HasSuffix(s, "}") {
			out[c] = v
			continue
		}

		name := s[2 : len(s)-1]
		env := os.Getenv(name)
		if env == "" {
			return nil, fmt.Errorf("column %q needs environment variable %s, which is unset", c, name)
		}

		// A password is never stored as given.
		if strings.Contains(strings.ToLower(c), "password") {
			hashed, err := security.Hash(env)
			if err != nil {
				return nil, err
			}
			out[c] = hashed
			continue
		}
		out[c] = env
	}
	return out, nil
}
