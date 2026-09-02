// Package domain holds the entities and the repository interfaces the
// application layer depends on.
//
// Nothing here imports GORM, database/sql, or net/http. That is the point: the
// application layer talks to these interfaces and cannot tell which engine is
// underneath, or whether there is a cache in front (SPEC.md §3).
package domain

import "errors"

var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("conflict")
	ErrInvalid  = errors.New("invalid")
)
