// Package dialect holds the engine-specific SQL that GORM does not render:
// full-text search and sequence reset.
package dialect

import (
	"context"

	"gorm.io/gorm"
)

// Dialect is injected into repositories. It must not appear in domain or
// application signatures.
type Dialect interface {
	// Name is "postgres" or "mysql".
	Name() string

	// FullTextCondition returns a WHERE fragment and its bind argument.
	FullTextCondition(columns []string, query string) (string, any)

	// ResetSequence sets the next generated id, required after importing rows
	// with explicit ids.
	ResetSequence(ctx context.Context, db *gorm.DB, table, idColumn string, next int64) error
}
