package cache

import (
	"context"
	"fmt"
	"strings"
	"time"

	"dpmptsp/api/internal/domain"
)

// Resource names a group of tables sharing one version counter.
type Resource string

const (
	// Branding and navigation: read on every page, written rarely.
	ResourceChrome Resource = "chrome"
	// Articles and their categories.
	ResourceArticles Resource = "articles"
	// The rarely-written content tables: regulasi, ppid, pelayanan, inovasi…
	ResourceContent Resource = "content"
)

// TTLs. Keys built from a superseded version expire on their own.
const (
	TTLChrome   = 24 * time.Hour
	TTLArticles = 10 * time.Minute
	TTLDetail   = 30 * time.Minute
	TTLContent  = time.Hour
)

// VersionedKey builds a cache key carrying the resource's current version.
// Incrementing the counter retires every key built from the previous one, so
// list invalidation needs no pattern delete.
func VersionedKey(ctx context.Context, c *Client, r Resource, parts ...string) (string, bool) {
	if c == nil {
		return "", false
	}
	v, err := c.Version(ctx, string(r))
	if err != nil {
		return "", false
	}
	return fmt.Sprintf("%s:v%d:%s", r, v, strings.Join(parts, ":")), true
}

// Invalidate increments a resource's version counter.
func Invalidate(ctx context.Context, c *Client, r Resource) {
	if c == nil {
		return
	}
	_ = c.BumpVersion(ctx, string(r))
}

// EntityKey addresses a single record. Such keys are deleted directly on write.
func EntityKey(r Resource, id string) string {
	return fmt.Sprintf("%s:entity:%s", r, id)
}

// ArticleListKey encodes every filter field that affects the result set.
func ArticleListKey(ctx context.Context, c *Client, f domain.ArticleFilter) (string, bool) {
	cats := make([]string, 0, len(f.CategoryIDs))
	for _, id := range f.CategoryIDs {
		cats = append(cats, fmt.Sprint(id))
	}
	return VersionedKey(ctx, c, ResourceArticles,
		"list",
		"cat="+strings.Join(cats, "."),
		fmt.Sprintf("active=%t", f.ActiveOnly),
		fmt.Sprintf("headline=%t", f.HeadlineOnly),
		fmt.Sprintf("limit=%d", f.Limit),
		fmt.Sprintf("offset=%d", f.Offset),
	)
}
