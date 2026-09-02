// Package dialect isolates the few things that genuinely differ between
// Postgres and MySQL and that GORM does not render for us.
//
// GORM already handles placeholders, RETURNING versus LastInsertId, and
// conflict clauses. Most of the rest was designed away rather than abstracted:
//
//	ENUM                        -> VARCHAR + CHECK, identical on both
//	ON UPDATE CURRENT_TIMESTAMP -> GORM's UpdatedAt
//	JSONB                       -> TEXT, decoded in Go (never queried by path)
//	partial indexes             -> full indexes
//	CURDATE()/NOW()             -> time.Now() as a bind parameter
//
// What is left is this: full-text search, which has no common syntax, and
// sequence reset, which is needed after importing rows with explicit ids.
package dialect

import (
	"context"

	"gorm.io/gorm"
)

// Dialect carries the engine-specific behaviour. It is injected into
// repositories and must never appear in a domain or application signature —
// those layers cannot know which engine is underneath (SPEC.md §3).
type Dialect interface {
	// Name is "postgres" or "mysql".
	Name() string

	// FullTextCondition returns a WHERE fragment and its argument for a
	// full-text search over the given columns. Postgres uses tsvector, MySQL
	// uses MATCH ... AGAINST; there is no shared syntax.
	FullTextCondition(columns []string, query string) (string, any)

	// ResetSequence sets the next generated id for a table. Required after a
	// bulk import with explicit ids, or the first insert collides.
	ResetSequence(ctx context.Context, db *gorm.DB, table, idColumn string, next int64) error
}
