package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// LocalDisk stores objects on the filesystem.
//
// With BaseURL empty the disk is private and URL returns ErrNotPublic;
// otherwise files are assumed to be served from that prefix.
type LocalDisk struct {
	name    string
	root    string
	baseURL string
}

func NewLocalDisk(name, root, baseURL string) (*LocalDisk, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("storage %s: resolve root: %w", name, err)
	}
	// The root is created lazily: every disk is constructed regardless of which
	// one is selected, so construction must not require a writable directory.
	return &LocalDisk{name: name, root: abs, baseURL: strings.TrimRight(baseURL, "/")}, nil
}

func (d *LocalDisk) Name() string { return d.name }

// dirMode decides whether other users may traverse into the disk.
//
// A disk with a base URL is served by another process — nginx, from a shared
// volume — which runs as a different user. Private directories made its files
// unreachable no matter what mode the files themselves carried.
func (d *LocalDisk) dirMode() os.FileMode {
	if d.baseURL != "" {
		return 0o755
	}
	return 0o750
}

// EnsureWritable creates the disk root and verifies it is writable. FromEnv
// calls this for the selected disk only.
func (d *LocalDisk) EnsureWritable() error {
	if err := os.MkdirAll(d.root, d.dirMode()); err != nil {
		return fmt.Errorf("storage %s: create root %s: %w", d.name, d.root, err)
	}
	probe, err := os.CreateTemp(d.root, ".writable-*")
	if err != nil {
		return fmt.Errorf("storage %s: root %s is not writable: %w", d.name, d.root, err)
	}
	name := probe.Name()
	probe.Close()
	return os.Remove(name)
}

// resolve maps a key to an absolute path inside the disk root.
func (d *LocalDisk) resolve(key string) (string, error) {
	clean, err := cleanKey(key)
	if err != nil {
		return "", err
	}
	full := filepath.Join(d.root, filepath.FromSlash(clean))
	if !strings.HasPrefix(full, d.root+string(os.PathSeparator)) {
		return "", fmt.Errorf("storage: key escapes the disk root: %q", key)
	}
	return full, nil
}

// chmodTree fixes the mode of directories created for key, up to the root.
func (d *LocalDisk) chmodTree(dir string) error {
	mode := d.dirMode()
	for len(dir) > len(d.root) {
		if err := os.Chmod(dir, mode); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("storage %s: chmod dir: %w", d.name, err)
		}
		dir = filepath.Dir(dir)
	}
	return os.Chmod(d.root, mode)
}

func (d *LocalDisk) Put(_ context.Context, key string, r io.Reader, opts PutOptions) (*Object, error) {
	full, err := d.resolve(key)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(full), d.dirMode()); err != nil {
		return nil, fmt.Errorf("storage %s: mkdir: %w", d.name, err)
	}
	// MkdirAll applies the umask, so a public disk's directories are chmod'd
	// back to the intended mode.
	if err := d.chmodTree(filepath.Dir(full)); err != nil {
		return nil, err
	}

	// Write to a temporary file and rename into place, so a partial write is
	// never observable.
	tmp, err := os.CreateTemp(filepath.Dir(full), ".upload-*")
	if err != nil {
		return nil, fmt.Errorf("storage %s: temp file: %w", d.name, err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	n, err := io.Copy(tmp, r)
	if err != nil {
		tmp.Close()
		return nil, fmt.Errorf("storage %s: write: %w", d.name, err)
	}
	if err := tmp.Close(); err != nil {
		return nil, fmt.Errorf("storage %s: close: %w", d.name, err)
	}

	mode := os.FileMode(0o640)
	if opts.Visibility == Public {
		mode = 0o644
	}
	if err := os.Chmod(tmpName, mode); err != nil {
		return nil, fmt.Errorf("storage %s: chmod: %w", d.name, err)
	}
	if err := os.Rename(tmpName, full); err != nil {
		return nil, fmt.Errorf("storage %s: rename: %w", d.name, err)
	}

	clean, _ := cleanKey(key)
	return &Object{Key: clean, Size: n, ContentType: opts.ContentType, Visibility: opts.Visibility}, nil
}

func (d *LocalDisk) Get(_ context.Context, key string) (io.ReadCloser, error) {
	full, err := d.resolve(key)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(full)
	if errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("%w: %s", ErrNotFound, key)
	}
	if err != nil {
		return nil, fmt.Errorf("storage %s: open: %w", d.name, err)
	}
	return f, nil
}

func (d *LocalDisk) Delete(_ context.Context, key string) error {
	full, err := d.resolve(key)
	if err != nil {
		return err
	}
	if err := os.Remove(full); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("storage %s: delete: %w", d.name, err)
	}
	return nil
}

func (d *LocalDisk) Exists(_ context.Context, key string) (bool, error) {
	full, err := d.resolve(key)
	if err != nil {
		return false, err
	}
	if _, err := os.Stat(full); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, fmt.Errorf("storage %s: stat: %w", d.name, err)
	}
	return true, nil
}

func (d *LocalDisk) URL(key string) (string, error) {
	if d.baseURL == "" {
		return "", fmt.Errorf("%w: %s", ErrNotPublic, d.name)
	}
	clean, err := cleanKey(key)
	if err != nil {
		return "", err
	}
	// Escape each segment; keys may contain spaces and non-ASCII characters.
	parts := strings.Split(clean, "/")
	for i, p := range parts {
		parts[i] = url.PathEscape(p)
	}
	return d.baseURL + "/" + path.Join(parts...), nil
}
