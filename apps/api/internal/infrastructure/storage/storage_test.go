package storage

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func newTestDisk(t *testing.T, baseURL string) *LocalDisk {
	t.Helper()
	d, err := NewLocalDisk("test", t.TempDir(), baseURL)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func TestLocalRoundTrip(t *testing.T) {
	d := newTestDisk(t, "")
	ctx := context.Background()

	obj, err := d.Put(ctx, "a/b/file.txt", strings.NewReader("hello"), PutOptions{ContentType: "text/plain"})
	if err != nil {
		t.Fatal(err)
	}
	if obj.Size != 5 {
		t.Errorf("size = %d, want 5", obj.Size)
	}

	rc, err := d.Get(ctx, "a/b/file.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer rc.Close()
	got, _ := io.ReadAll(rc)
	if string(got) != "hello" {
		t.Errorf("content = %q", got)
	}

	if ok, _ := d.Exists(ctx, "a/b/file.txt"); !ok {
		t.Error("Exists should be true")
	}
	if err := d.Delete(ctx, "a/b/file.txt"); err != nil {
		t.Fatal(err)
	}
	if ok, _ := d.Exists(ctx, "a/b/file.txt"); ok {
		t.Error("Exists should be false after delete")
	}
}

func TestDeletingSomethingAbsentIsNotAnError(t *testing.T) {
	// Callers are usually cleaning up and should not have to check first.
	if err := newTestDisk(t, "").Delete(context.Background(), "never/existed"); err != nil {
		t.Errorf("got %v", err)
	}
}

func TestGetMissingReportsNotFound(t *testing.T) {
	_, err := newTestDisk(t, "").Get(context.Background(), "nope.txt")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("got %v, want ErrNotFound", err)
	}
}

// The bug this whole package replaces: the Astro upload routes joined the raw
// client filename onto the upload directory, so a crafted name wrote outside it.
func TestKeysCannotEscapeTheDiskRoot(t *testing.T) {
	root := t.TempDir()
	d, err := NewLocalDisk("test", root, "")
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	for _, key := range []string{
		"../escaped.txt",
		"../../escaped.txt",
		"a/../../escaped.txt",
		`..\..\escaped.txt`,
		"/etc/passwd",
	} {
		_, err := d.Put(ctx, key, strings.NewReader("x"), PutOptions{})
		if err == nil {
			// A key that resolves back inside the root is acceptable; one that
			// lands outside it is not.
			outside := filepath.Join(filepath.Dir(root), "escaped.txt")
			if _, statErr := os.Stat(outside); statErr == nil {
				t.Fatalf("key %q wrote outside the disk root", key)
			}
			continue
		}
	}
}

func TestPrivateDiskHasNoURL(t *testing.T) {
	_, err := newTestDisk(t, "").URL("x.png")
	if !errors.Is(err, ErrNotPublic) {
		t.Errorf("got %v, want ErrNotPublic", err)
	}
}

func TestPublicDiskURLEscapesSegments(t *testing.T) {
	d := newTestDisk(t, "https://cdn.example/storage/")
	got, err := d.URL("artikel/2026/01/foto bagus.jpg")
	if err != nil {
		t.Fatal(err)
	}
	want := "https://cdn.example/storage/artikel/2026/01/foto%20bagus.jpg"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// --- upload policy ---

// A PNG magic header. The point is that the declared filename says .txt and the
// bytes say PNG; the bytes must win.
var pngBytes = append([]byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a}, bytes.Repeat([]byte{0}, 64)...)

func TestUploadSniffsTypeRatherThanTrustingTheFilename(t *testing.T) {
	d := newTestDisk(t, "")
	obj, err := Upload(context.Background(), d, bytes.NewReader(pngBytes), "notes.txt",
		UploadRules{Prefix: "artikel"})
	if err != nil {
		t.Fatal(err)
	}
	if obj.ContentType != "image/png" {
		t.Errorf("content type = %q, want image/png", obj.ContentType)
	}
	if !strings.HasSuffix(obj.Key, ".png") {
		t.Errorf("key %q should take its extension from the sniffed type", obj.Key)
	}
}

func TestUploadRejectsADisallowedType(t *testing.T) {
	d := newTestDisk(t, "")
	_, err := Upload(context.Background(), d, strings.NewReader("<html>hi</html>"), "x.png",
		UploadRules{AllowedTypes: []string{"image/png"}})
	if !errors.Is(err, ErrTypeRejected) {
		t.Errorf("got %v, want ErrTypeRejected", err)
	}
}

func TestUploadEnforcesTheSizeCapAndCleansUp(t *testing.T) {
	d := newTestDisk(t, "")
	big := append(pngBytes, bytes.Repeat([]byte{1}, 4096)...)

	_, err := Upload(context.Background(), d, bytes.NewReader(big), "big.png",
		UploadRules{MaxBytes: 1024, AllowedTypes: []string{"image/png"}})
	if !errors.Is(err, ErrTooLarge) {
		t.Fatalf("got %v, want ErrTooLarge", err)
	}

	// The rejected object must not be left behind.
	root := filepath.Join(d.root)
	var found []string
	_ = filepath.Walk(root, func(p string, info os.FileInfo, _ error) error {
		if info != nil && !info.IsDir() {
			found = append(found, p)
		}
		return nil
	})
	if len(found) != 0 {
		t.Errorf("oversized upload left files behind: %v", found)
	}
}

func TestUploadKeysAreUnique(t *testing.T) {
	d := newTestDisk(t, "")
	ctx := context.Background()
	seen := map[string]bool{}

	// The old scheme used Date.now(), so two uploads in the same millisecond
	// silently overwrote each other.
	for i := 0; i < 200; i++ {
		obj, err := Upload(ctx, d, bytes.NewReader(pngBytes), "same-name.png", UploadRules{})
		if err != nil {
			t.Fatal(err)
		}
		if seen[obj.Key] {
			t.Fatalf("duplicate key generated: %s", obj.Key)
		}
		seen[obj.Key] = true
	}
}

func TestManagerResolvesDisks(t *testing.T) {
	local := newTestDisk(t, "")
	pub := newTestDisk(t, "https://cdn.example")

	m, err := NewManager("local", map[string]Disk{"local": local, "public": pub})
	if err != nil {
		t.Fatal(err)
	}
	if m.Default().Name() != "test" {
		t.Errorf("default disk = %q", m.Default().Name())
	}
	if _, err := m.Disk("public"); err != nil {
		t.Errorf("public disk should resolve: %v", err)
	}
	if _, err := m.Disk("s3"); !errors.Is(err, ErrUnknownDisk) {
		t.Errorf("got %v, want ErrUnknownDisk", err)
	}
	// An empty name means "the default", which is what makes FILESYSTEM_DISK
	// the single switch.
	if d, err := m.Disk(""); err != nil || d.Name() != "test" {
		t.Errorf("empty name should resolve to the default, got %v %v", d, err)
	}
}

func TestManagerRejectsAnUnconfiguredDefault(t *testing.T) {
	if _, err := NewManager("s3", map[string]Disk{"local": newTestDisk(t, "")}); !errors.Is(err, ErrUnknownDisk) {
		t.Errorf("got %v, want ErrUnknownDisk", err)
	}
}
