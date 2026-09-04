package cache

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"
)

// Every write on a caching decorator must retire what it invalidates.
//
// The rule is easy to break by adding one more Create/Update/Delete that just
// forwards to the inner repository: reads keep serving the old value until the
// TTL runs out, and nothing fails. This walks the decorators' own source and
// checks each write method reaches an invalidation.
func TestEveryCachedWriteInvalidates(t *testing.T) {
	files, err := filepath.Glob("cached_*_repo.go")
	if err != nil || len(files) == 0 {
		t.Fatalf("no decorator sources found: %v", err)
	}

	fset := token.NewFileSet()
	for _, file := range files {
		f, err := parser.ParseFile(fset, file, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", file, err)
		}

		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv == nil || fn.Body == nil {
				continue
			}
			if !isWrite(fn.Name.Name) {
				continue
			}
			if !callsInvalidation(fn.Body) {
				t.Errorf("%s: %s writes without invalidating anything", file, fn.Name.Name)
			}
		}
	}
}

func isWrite(name string) bool {
	for _, prefix := range []string{"Create", "Update", "Delete"} {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}

// callsInvalidation accepts either a direct Invalidate/Del call or a hop
// through the decorator's own invalidate helper.
func callsInvalidation(body *ast.BlockStmt) bool {
	found := false
	ast.Inspect(body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		switch fn := call.Fun.(type) {
		case *ast.Ident:
			if fn.Name == "Invalidate" || fn.Name == "Del" {
				found = true
			}
		case *ast.SelectorExpr:
			if fn.Sel.Name == "invalidate" || fn.Sel.Name == "Invalidate" {
				found = true
			}
		}
		return true
	})
	return found
}

// The homepage embeds articles, announcements and content, so it carries its
// own counter. Keying it under any one of them leaves it stale when either of
// the others is written.
func TestHomeHasItsOwnResource(t *testing.T) {
	for _, r := range []Resource{ResourceArticles, ResourceContent, ResourceChrome} {
		if ResourceHome == r {
			t.Fatalf("ResourceHome must not share a counter with %q", r)
		}
	}
}
