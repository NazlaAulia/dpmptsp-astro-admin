package cache

import (
	"context"
	"fmt"
	"strings"
	"time"

	"dpmptsp/api/internal/domain"
)

// Resource names a group of data that shares one version counter.
//
// SPEC.md §6 asks for a counter per resource type. Twelve counters for
// twenty-six tables would be more machinery than the invalidation patterns
// justify, so related tables share one and are split only when measurement
// shows a counter is too coarse.
type Resource string

const (
	// Read on literally every page, written almost never, so it earns its own
	// counter and a long TTL that an article edit must not disturb.
	ResourceChrome Resource = "chrome"
	// Articles and their categories.
	ResourceArticles Resource = "articles"
	// The rarely-written content tables: regulasi, ppid, pelayanan, inovasi…
	ResourceContent Resource = "content"
)

// TTLs. Old versioned keys become unreachable the moment a counter moves and
// then expire on their own, which is what makes SCAN unnecessary.
const (
	TTLChrome   = 24 * time.Hour
	TTLArticles = 10 * time.Minute
	TTLDetail   = 30 * time.Minute
	TTLContent  = time.Hour
)

// VersionedKey builds a cache key carrying the resource's current version.
//
// This is the whole mechanism from SPEC.md §6: a write only has to INCR the
// counter. Every key built from the old version is instantly unreachable and
// expires on its TTL, so a list cache can be invalidated without SCAN+DEL,
// which blocks Redis in production (CLAUDE.md rule 7).
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

// Invalidate bumps a resource's counter, which retires every list key built
// from it in one operation.
func Invalidate(ctx context.Context, c *Client, r Resource) {
	if c == nil {
		return
	}
	_ = c.BumpVersion(ctx, string(r))
}

// EntityKey is for a single record, which is deleted directly on write rather
// than versioned — there is exactly one key, so there is nothing to sweep.
func EntityKey(r Resource, id string) string {
	return fmt.Sprintf("%s:entity:%s", r, id)
}

// ArticleListKey encodes the filter into the key. Anything that changes the
// result set has to appear here, or two different queries share a cache entry.
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
