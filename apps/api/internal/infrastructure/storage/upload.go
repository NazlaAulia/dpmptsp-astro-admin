package storage

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"time"
)

// DefaultMaxUploadBytes is the default per-upload size limit.
const DefaultMaxUploadBytes int64 = 50 << 20 // 50 MB

// DefaultAllowedTypes is the default content-type allowlist.
var DefaultAllowedTypes = []string{
	"image/jpeg", "image/png", "image/webp", "image/gif",
	"application/pdf",
	"video/mp4", "video/webm", "video/quicktime",
}

// UploadRules is the policy applied to an incoming file.
type UploadRules struct {
	MaxBytes     int64
	AllowedTypes []string
	// Prefix is prepended to the generated key, e.g. "artikel".
	Prefix     string
	Visibility Visibility
}

func (r UploadRules) withDefaults() UploadRules {
	if r.MaxBytes <= 0 {
		r.MaxBytes = DefaultMaxUploadBytes
	}
	if len(r.AllowedTypes) == 0 {
		r.AllowedTypes = DefaultAllowedTypes
	}
	if r.Visibility == "" {
		r.Visibility = Private
	}
	return r
}

// Upload validates and stores an incoming file.
//
// The content type is detected from the object's leading bytes rather than the
// client's declaration, the size limit is applied while streaming, and the
// stored key is generated rather than derived from the supplied filename.
func Upload(ctx context.Context, disk Disk, r io.Reader, filename string, rules UploadRules) (*Object, error) {
	rules = rules.withDefaults()

	// Detect the type from the leading bytes, then replay them into the body.
	header := make([]byte, 512)
	n, err := io.ReadFull(r, header)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return nil, fmt.Errorf("storage: read header: %w", err)
	}
	header = header[:n]
	contentType := normaliseType(http.DetectContentType(header))

	if !allowed(contentType, rules.AllowedTypes) {
		return nil, fmt.Errorf("%w: %s", ErrTypeRejected, contentType)
	}

	body := io.MultiReader(strings.NewReader(string(header)), r)
	// One byte over the cap, so an oversized upload is detectable.
	limited := io.LimitReader(body, rules.MaxBytes+1)

	key, err := generateKey(rules.Prefix, filename, contentType)
	if err != nil {
		return nil, err
	}

	obj, err := disk.Put(ctx, key, limited, PutOptions{
		ContentType: contentType,
		Visibility:  rules.Visibility,
	})
	if err != nil {
		return nil, err
	}

	if obj.Size > rules.MaxBytes {
		// Remove the partial object.
		_ = disk.Delete(ctx, obj.Key)
		return nil, fmt.Errorf("%w: %d bytes, limit %d", ErrTooLarge, obj.Size, rules.MaxBytes)
	}
	return obj, nil
}

// generateKey builds a random, date-sharded object key.
func generateKey(prefix, filename, contentType string) (string, error) {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", fmt.Errorf("storage: generate key: %w", err)
	}

	// The extension follows the detected type.
	ext := extensionFor(contentType)
	if ext == ".bin" {
		// Unrecognised type: fall back to the supplied extension if it is
		// short and alphanumeric.
		if candidate := strings.ToLower(path.Ext(filename)); safeExt(candidate) {
			ext = candidate
		}
	}

	name := hex.EncodeToString(buf[:]) + ext
	// Shard by date to bound directory size.
	datePath := time.Now().UTC().Format("2006/01")

	if prefix != "" {
		return path.Join(prefix, datePath, name), nil
	}
	return path.Join(datePath, name), nil
}

func safeExt(ext string) bool {
	if len(ext) < 2 || len(ext) > 6 {
		return false
	}
	for _, c := range ext[1:] {
		if !(c >= 'a' && c <= 'z') && !(c >= '0' && c <= '9') {
			return false
		}
	}
	return true
}

func extensionFor(contentType string) string {
	switch contentType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	case "application/pdf":
		return ".pdf"
	case "video/mp4":
		return ".mp4"
	case "video/webm":
		return ".webm"
	case "video/quicktime":
		return ".mov"
	default:
		return ".bin"
	}
}

// normaliseType drops any charset parameter.
func normaliseType(t string) string {
	if i := strings.IndexByte(t, ';'); i >= 0 {
		return strings.TrimSpace(t[:i])
	}
	return t
}

func allowed(contentType string, list []string) bool {
	for _, a := range list {
		if strings.EqualFold(a, contentType) {
			return true
		}
	}
	return false
}
