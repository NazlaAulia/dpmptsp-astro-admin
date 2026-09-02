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

// DefaultMaxUploadBytes matches the cap the old Astro endpoint enforced.
const DefaultMaxUploadBytes int64 = 50 << 20 // 50 MB

// DefaultAllowedTypes is the allowlist carried over from the endpoint that had
// one. Everything else uploaded through this API had no validation at all.
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
// Three things the Astro upload routes got wrong, fixed here:
//
//  1. The content type is sniffed from the first bytes, not taken from the
//     client's Content-Type. A client can claim anything; a .jpg that is really
//     an HTML document gets stored and later served.
//
//  2. The size limit is enforced while streaming, via io.LimitReader, rather
//     than after the body is in memory. Checking afterwards means a 2 GB upload
//     costs 2 GB of RAM before being rejected.
//
//  3. The stored key is generated. The client's filename is used for nothing
//     but a sanity-checked extension — the old code passed it straight to
//     path.Join, which is a path traversal, and its Date.now() naming collided
//     whenever two uploads landed in the same millisecond.
func Upload(ctx context.Context, disk Disk, r io.Reader, filename string, rules UploadRules) (*Object, error) {
	rules = rules.withDefaults()

	// Sniff from the leading bytes, then put them back so nothing is consumed.
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
	// One byte over the cap so exceeding it is detectable rather than silently
	// truncating the file.
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
		// Do not leave the oversized object behind.
		_ = disk.Delete(ctx, obj.Key)
		return nil, fmt.Errorf("%w: %d bytes, limit %d", ErrTooLarge, obj.Size, rules.MaxBytes)
	}
	return obj, nil
}

// generateKey builds an opaque, collision-free key. Only the extension is
// derived from the client, and only when it is short and alphanumeric.
func generateKey(prefix, filename, contentType string) (string, error) {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", fmt.Errorf("storage: generate key: %w", err)
	}

	// The extension comes from what the bytes ARE, not from what the upload
	// claimed to be. Preferring the client's extension would store a PNG as
	// notes.txt purely because that is what the browser sent, which is the
	// same "trust the client" mistake this package exists to avoid.
	ext := extensionFor(contentType)
	if ext == ".bin" {
		// Type not recognised: fall back to the client's extension, but only
		// if it is short and alphanumeric.
		if candidate := strings.ToLower(path.Ext(filename)); safeExt(candidate) {
			ext = candidate
		}
	}

	name := hex.EncodeToString(buf[:]) + ext
	// Date-shard so no single directory accumulates every object ever uploaded.
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

// normaliseType drops the charset parameter DetectContentType adds to text.
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
