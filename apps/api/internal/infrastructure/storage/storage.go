// Package storage provides file storage over swappable disks.
//
// Disks are configured by environment and selected by name, so the backing
// store can change without touching call sites.
package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"
)

var (
	ErrNotFound     = errors.New("storage: object not found")
	ErrUnknownDisk  = errors.New("storage: unknown disk")
	ErrNotPublic    = errors.New("storage: disk has no public URL")
	ErrTooLarge     = errors.New("storage: object exceeds the size limit")
	ErrTypeRejected = errors.New("storage: content type not allowed")
)

// Visibility controls whether an object is reachable by URL. Private objects
// are served only through an authorizing handler.
type Visibility string

const (
	Private Visibility = "private"
	Public  Visibility = "public"
)

// Object describes something that was stored.
type Object struct {
	// Key is the path within the disk, e.g. "artikel/2026/abc123.jpg".
	Key string
	// Size in bytes, as actually written.
	Size int64
	// ContentType detected from the object's bytes.
	ContentType string
	// Visibility the object was stored with.
	Visibility Visibility
}

// PutOptions carries everything a write needs beyond the bytes themselves.
type PutOptions struct {
	ContentType string
	Visibility  Visibility
}

// Disk is one configured storage location.
//
// Backend-specific operations such as presigned URLs are not part of this
// interface; reach them on the concrete type.
type Disk interface {
	// Name is the configured disk name, for logs and errors.
	Name() string

	// Put writes r under key and reports what was stored, including the number
	// of bytes actually written.
	Put(ctx context.Context, key string, r io.Reader, opts PutOptions) (*Object, error)

	// Get opens an object for reading. The caller closes it.
	Get(ctx context.Context, key string) (io.ReadCloser, error)

	// Delete removes an object. Deleting an absent object is not an error.
	Delete(ctx context.Context, key string) error

	// Exists reports whether an object is present.
	Exists(ctx context.Context, key string) (bool, error)

	// URL returns a public URL, or ErrNotPublic if the disk does not serve one.
	URL(key string) (string, error)
}

// Manager resolves configured disks by name.
type Manager struct {
	disks       map[string]Disk
	defaultDisk string
}

func NewManager(defaultDisk string, disks map[string]Disk) (*Manager, error) {
	if _, ok := disks[defaultDisk]; !ok {
		return nil, fmt.Errorf("%w: default disk %q is not configured", ErrUnknownDisk, defaultDisk)
	}
	return &Manager{disks: disks, defaultDisk: defaultDisk}, nil
}

// Disk returns the named disk. An empty name returns the default.
func (m *Manager) Disk(name string) (Disk, error) {
	if name == "" {
		name = m.defaultDisk
	}
	d, ok := m.disks[name]
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrUnknownDisk, name)
	}
	return d, nil
}

// Default returns the disk named by FILESYSTEM_DISK.
func (m *Manager) Default() Disk {
	return m.disks[m.defaultDisk]
}

// Names lists the configured disks.
func (m *Manager) Names() []string {
	out := make([]string, 0, len(m.disks))
	for n := range m.disks {
		out = append(out, n)
	}
	return out
}

// cleanKey normalises an object key and rejects paths that escape the disk root.
func cleanKey(key string) (string, error) {
	k := strings.TrimPrefix(path.Clean("/"+strings.ReplaceAll(key, `\`, "/")), "/")
	if k == "" || k == "." {
		return "", fmt.Errorf("storage: empty key")
	}
	if strings.HasPrefix(k, "../") || strings.Contains(k, "/../") {
		return "", fmt.Errorf("storage: key escapes the disk root: %q", key)
	}
	return k, nil
}
