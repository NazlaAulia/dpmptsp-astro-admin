// Package cache holds the Redis client and the version-counter helpers that
// SPEC.md §6 specifies for list invalidation.
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	rdb *redis.Client
}

func Open(ctx context.Context, url string) (*Client, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	rdb := redis.NewClient(opts)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := rdb.Ping(pingCtx).Err(); err != nil {
		rdb.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}
	return &Client{rdb: rdb}, nil
}

func (c *Client) Close() error { return c.rdb.Close() }

func (c *Client) Ping(ctx context.Context) error { return c.rdb.Ping(ctx).Err() }

// Version returns the current version counter for a resource type. Cache keys
// embed it, so a write only has to bump the counter: old keys become orphans
// and expire on their own TTL.
//
// This is what lets list caches be invalidated without SCAN, which blocks the
// server and is forbidden by CLAUDE.md rule 7.
func (c *Client) Version(ctx context.Context, resource string) (int64, error) {
	key := resource + ":version"
	n, err := c.rdb.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("read %s: %w", key, err)
	}
	return n, nil
}

// BumpVersion is called on every write to a resource type.
func (c *Client) BumpVersion(ctx context.Context, resource string) error {
	key := resource + ":version"
	if err := c.rdb.Incr(ctx, key).Err(); err != nil {
		return fmt.Errorf("incr %s: %w", key, err)
	}
	return nil
}

// --- typed helpers -----------------------------------------------------------

// GetJSON reads and decodes a cached value. A miss, a decode failure, or an
// unreachable Redis all report "not cached" rather than an error: the cache is
// an optimisation (SPEC.md §6), and a broken cache must degrade to a slower
// request, never to a failed one.
func GetJSON[T any](ctx context.Context, c *Client, key string) (T, bool) {
	var zero T
	if c == nil {
		return zero, false
	}
	raw, err := c.rdb.Get(ctx, key).Bytes()
	if err != nil {
		return zero, false
	}
	var out T
	if err := json.Unmarshal(raw, &out); err != nil {
		return zero, false
	}
	return out, true
}

// SetJSON stores a value with a TTL. Failures are ignored for the same reason.
func SetJSON(ctx context.Context, c *Client, key string, v any, ttl time.Duration) {
	if c == nil {
		return
	}
	raw, err := json.Marshal(v)
	if err != nil {
		return
	}
	_ = c.rdb.Set(ctx, key, raw, ttl).Err()
}

// Del removes specific keys. Only ever called with keys we constructed — never
// a pattern, because SCAN blocks the server (CLAUDE.md rule 7).
func Del(ctx context.Context, c *Client, keys ...string) {
	if c == nil || len(keys) == 0 {
		return
	}
	_ = c.rdb.Del(ctx, keys...).Err()
}
