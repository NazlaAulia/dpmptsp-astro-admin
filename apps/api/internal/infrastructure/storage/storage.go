// Package storage provides file storage with swappable drivers, in the shape
// Laravel's filesystem takes: code asks for a disk by name and never knows
// whether it is writing to local disk or to an S3 bucket.
//
// Disks are configured by environment, so moving uploads from a mounted volume
// to MinIO is a configuration change, not a code change. That matters here
// because uploads currently land in the Astro apps' public/ directories, which
// cannot survive more than one replica.
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

// Visibility mirrors Laravel's public/private distinction. A private object is
// only reachable through a handler that can check authorization; a public one
// has a URL anybody can fetch.
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
	// ContentType as detected from the bytes, not as claimed by the client.
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
// Deliberately small. Everything an upload handler needs is here, and nothing
// that would tie the interface to one backend — no presigned-URL method, no
// bucket policy, no directory listing. Those can be added to the concrete S3
// implementation and reached through a type assertion when something genuinely
// needs them.
type Disk interface {
	// Name is the configured disk name, for logs and errors.
	Name() string

	// Put writes r under key. It returns what was actually stored, including
	// the size written, so a caller does not have to trust a Content-Length it
	// was given.
	Put(ctx context.Context, key string, r io.Reader, opts PutOptions) (*Object, error)

	// Get opens an object for reading. The caller closes it.
	Get(ctx context.Context, key string) (io.ReadCloser, error)

	// Delete removes an object. Deleting something that is not there is not an
	// error: callers are usually cleaning up and should not have to care.
	Delete(ctx context.Context, key string) error

	// Exists reports whether an object is present.
	Exists(ctx context.Context, key string) (bool, error)

	// URL returns a publicly fetchable URL, or ErrNotPublic if this disk does
	// not serve one. Private disks are served through an authorizing handler
	// instead.
	URL(key string) (string, error)
}

// Manager resolves disks by name, so call sites read like
// storage.Disk("s3") rather than threading a concrete type through.
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

// Disk returns the named disk. An empty name returns the default, which is what
// makes FILESYSTEM_DISK the single switch.
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

// Names lists the configured disks, for the health endpoint and for errors.
func (m *Manager) Names() []string {
	out := make([]string, 0, len(m.disks))
	for n := range m.disks {
		out = append(out, n)
	}
	return out
}

// cleanKey normalises an object key and refuses anything that tries to escape
// the disk root.
//
// This is the guard the current Astro upload routes lack: they join the raw
// client-supplied filename onto the upload directory, so a multipart filename of
// "../../src/pages/index.astro" writes outside it.
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
