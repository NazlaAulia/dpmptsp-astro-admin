package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// S3Disk stores objects in any S3-compatible bucket.
type S3Disk struct {
	name    string
	client  *minio.Client
	bucket  string
	baseURL string
}

type S3Config struct {
	Name      string
	Endpoint  string // host:port, no scheme
	Region    string
	Bucket    string
	AccessKey string
	SecretKey string
	UseSSL    bool
	// BaseURL, when set, is the public prefix objects are served from — a CDN
	// or a public bucket. Empty means the disk is private and URL() refuses.
	BaseURL string
}

func NewS3Disk(ctx context.Context, cfg S3Config) (*S3Disk, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("storage %s: client: %w", cfg.Name, err)
	}

	// Fail at startup rather than on the first upload. A missing bucket or bad
	// credentials should stop the container, not surface as a 500 later.
	ok, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("storage %s: reach bucket %q: %w", cfg.Name, cfg.Bucket, err)
	}
	if !ok {
		return nil, fmt.Errorf("storage %s: bucket %q does not exist", cfg.Name, cfg.Bucket)
	}

	return &S3Disk{
		name:    cfg.Name,
		client:  client,
		bucket:  cfg.Bucket,
		baseURL: strings.TrimRight(cfg.BaseURL, "/"),
	}, nil
}

func (d *S3Disk) Name() string { return d.name }

func (d *S3Disk) Put(ctx context.Context, key string, r io.Reader, opts PutOptions) (*Object, error) {
	clean, err := cleanKey(key)
	if err != nil {
		return nil, err
	}

	// -1 streams with a multipart upload rather than requiring the length up
	// front, so an upload never has to be buffered just to be measured.
	info, err := d.client.PutObject(ctx, d.bucket, clean, r, -1, minio.PutObjectOptions{
		ContentType: opts.ContentType,
	})
	if err != nil {
		return nil, fmt.Errorf("storage %s: put %q: %w", d.name, clean, err)
	}

	return &Object{
		Key:         clean,
		Size:        info.Size,
		ContentType: opts.ContentType,
		Visibility:  opts.Visibility,
	}, nil
}

func (d *S3Disk) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	clean, err := cleanKey(key)
	if err != nil {
		return nil, err
	}
	obj, err := d.client.GetObject(ctx, d.bucket, clean, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("storage %s: get %q: %w", d.name, clean, err)
	}
	// GetObject is lazy: it does not contact the server until first read, so a
	// missing object would otherwise surface at an awkward moment.
	if _, err := obj.Stat(); err != nil {
		obj.Close()
		if isNotFound(err) {
			return nil, fmt.Errorf("%w: %s", ErrNotFound, clean)
		}
		return nil, fmt.Errorf("storage %s: stat %q: %w", d.name, clean, err)
	}
	return obj, nil
}

func (d *S3Disk) Delete(ctx context.Context, key string) error {
	clean, err := cleanKey(key)
	if err != nil {
		return err
	}
	if err := d.client.RemoveObject(ctx, d.bucket, clean, minio.RemoveObjectOptions{}); err != nil {
		if isNotFound(err) {
			return nil
		}
		return fmt.Errorf("storage %s: delete %q: %w", d.name, clean, err)
	}
	return nil
}

func (d *S3Disk) Exists(ctx context.Context, key string) (bool, error) {
	clean, err := cleanKey(key)
	if err != nil {
		return false, err
	}
	if _, err := d.client.StatObject(ctx, d.bucket, clean, minio.StatObjectOptions{}); err != nil {
		if isNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("storage %s: stat %q: %w", d.name, clean, err)
	}
	return true, nil
}

func (d *S3Disk) URL(key string) (string, error) {
	if d.baseURL == "" {
		return "", fmt.Errorf("%w: %s", ErrNotPublic, d.name)
	}
	clean, err := cleanKey(key)
	if err != nil {
		return "", err
	}
	parts := strings.Split(clean, "/")
	for i, p := range parts {
		parts[i] = url.PathEscape(p)
	}
	return d.baseURL + "/" + path.Join(parts...), nil
}

// PresignedURL gives temporary access to a private object.
func (d *S3Disk) PresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	clean, err := cleanKey(key)
	if err != nil {
		return "", err
	}
	u, err := d.client.PresignedGetObject(ctx, d.bucket, clean, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("storage %s: presign %q: %w", d.name, clean, err)
	}
	return u.String(), nil
}

func isNotFound(err error) bool {
	var resp minio.ErrorResponse
	if errors.As(err, &resp) {
		return resp.Code == "NoSuchKey" || resp.Code == "NoSuchBucket" || resp.StatusCode == 404
	}
	return false
}
