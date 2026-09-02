package domain

import "context"

// SiteSettings is the site's branding. One row, looked up by id 1.
type SiteSettings struct {
	Name string
	Logo string
}

// MenuNode is one entry of the header navigation, already nested.
//
// The tree is built in the API, not in the frontend. Reshaping flat parent_id
// rows into a tree is data work, and doing it in an Astro component meant every
// page render repeated it.
type MenuNode struct {
	ID            int64
	Name          string
	URL           string
	Order         int
	Type          string
	ContactButton bool
	Children      []MenuNode
}

// SiteChrome is everything the header and footer need, in one object, so a page
// makes one call rather than two.
type SiteChrome struct {
	Settings   SiteSettings
	Navigation []MenuNode
	Contact    []MenuNode
}

type SiteRepository interface {
	Settings(ctx context.Context) (SiteSettings, error)
	MenuTree(ctx context.Context) (nav []MenuNode, contact []MenuNode, err error)
}
