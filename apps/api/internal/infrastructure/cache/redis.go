// Package cache holds the Redis client and version-counter helpers used for
// list invalidation.
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

// Version returns the current counter for a resource type. Cache keys embed it,
// so incrementing the counter retires every key built from the previous value.
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

// BumpVersion increments a resource's counter.
func (c *Client) BumpVersion(ctx context.Context, resource string) error {
	key := resource + ":version"
	if err := c.rdb.Incr(ctx, key).Err(); err != nil {
		return fmt.Errorf("incr %s: %w", key, err)
	}
	return nil
}

// --- typed helpers -----------------------------------------------------------

// GetJSON reads and decodes a cached value. A miss, a decode failure and an
// unreachable server all report "not cached".
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

// SetJSON stores a value with a TTL. Failures are ignored.
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

// Del removes specific keys. Patterns are never used.
func Del(ctx context.Context, c *Client, keys ...string) {
	if c == nil || len(keys) == 0 {
		return
	}
	_ = c.rdb.Del(ctx, keys...).Err()
}
