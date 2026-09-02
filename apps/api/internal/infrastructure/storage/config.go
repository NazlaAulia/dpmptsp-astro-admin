package storage

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// FromEnv builds the disk registry, in the shape Laravel's filesystems config
// takes: several disks defined, one selected by FILESYSTEM_DISK.
//
//	FILESYSTEM_DISK       local | public | s3        (default: local)
//
//	STORAGE_LOCAL_ROOT    default /var/lib/dpmptsp/storage   (private)
//	STORAGE_PUBLIC_ROOT   default /var/lib/dpmptsp/public
//	STORAGE_PUBLIC_URL    default /storage                   (public prefix)
//
//	S3_ENDPOINT           host:port, no scheme. The s3 disk is only configured
//	                      when this is set, so a local-only deployment needs no
//	                      S3 variables at all.
//	S3_BUCKET S3_ACCESS_KEY S3_SECRET_KEY S3_REGION S3_USE_SSL S3_PUBLIC_URL
func FromEnv(ctx context.Context) (*Manager, error) {
	disks := map[string]Disk{}

	local, err := NewLocalDisk("local", envOr("STORAGE_LOCAL_ROOT", "/var/lib/dpmptsp/storage"), "")
	if err != nil {
		return nil, err
	}
	disks["local"] = local

	public, err := NewLocalDisk(
		"public",
		envOr("STORAGE_PUBLIC_ROOT", "/var/lib/dpmptsp/public"),
		envOr("STORAGE_PUBLIC_URL", "/storage"),
	)
	if err != nil {
		return nil, err
	}
	disks["public"] = public

	if endpoint := os.Getenv("S3_ENDPOINT"); endpoint != "" {
		s3, err := NewS3Disk(ctx, S3Config{
			Name:      "s3",
			Endpoint:  endpoint,
			Region:    envOr("S3_REGION", "us-east-1"),
			Bucket:    os.Getenv("S3_BUCKET"),
			AccessKey: os.Getenv("S3_ACCESS_KEY"),
			SecretKey: os.Getenv("S3_SECRET_KEY"),
			UseSSL:    strings.EqualFold(os.Getenv("S3_USE_SSL"), "true"),
			BaseURL:   os.Getenv("S3_PUBLIC_URL"),
		})
		if err != nil {
			return nil, err
		}
		disks["s3"] = s3
	}

	selected := envOr("FILESYSTEM_DISK", "local")
	if _, ok := disks[selected]; !ok {
		return nil, fmt.Errorf(
			"%w: FILESYSTEM_DISK=%q but only %s are configured (set S3_ENDPOINT to enable s3)",
			ErrUnknownDisk, selected, strings.Join(namesOf(disks), ", "))
	}
	// Verify only the disk actually in use.
	if l, ok := disks[selected].(*LocalDisk); ok {
		if err := l.EnsureWritable(); err != nil {
			return nil, err
		}
	}

	return NewManager(selected, disks)
}

func namesOf(disks map[string]Disk) []string {
	out := make([]string, 0, len(disks))
	for n := range disks {
		out = append(out, n)
	}
	return out
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
