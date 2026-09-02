package dialect

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// Postgres is the primary target (SPEC.md §12 as amended).
type Postgres struct{}

func (Postgres) Name() string { return "postgres" }

// FullTextCondition builds a tsvector predicate. coalesce matters: without it a
// NULL column nulls the whole document and the row silently stops matching.
func (Postgres) FullTextCondition(columns []string, query string) (string, any) {
	parts := make([]string, 0, len(columns))
	for _, c := range columns {
		parts = append(parts, fmt.Sprintf("coalesce(%s, '')", c))
	}
	return fmt.Sprintf(
		"to_tsvector('simple', %s) @@ plainto_tsquery('simple', ?)",
		strings.Join(parts, " || ' ' || "),
	), query
}

func (Postgres) ResetSequence(ctx context.Context, db *gorm.DB, table, idColumn string, next int64) error {
	// pg_get_serial_sequence resolves the sequence behind an identity column,
	// so this works without hardcoding sequence names.
	const stmt = `SELECT setval(pg_get_serial_sequence(?, ?), ?, false)`
	if err := db.WithContext(ctx).Exec(stmt, table, idColumn, next).Error; err != nil {
		return fmt.Errorf("reset sequence for %s.%s: %w", table, idColumn, err)
	}
	return nil
}
