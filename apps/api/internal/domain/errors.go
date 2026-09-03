// Package domain holds the entities and the repository interfaces the
// application layer depends on.
package domain

import "errors"

var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("conflict")
	ErrInvalid  = errors.New("invalid")
)
