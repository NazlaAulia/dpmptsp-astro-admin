package dialect

import (
	"strings"
	"testing"
)

// Both implementations must satisfy the interface. This compile-time guard is
// what stops one engine quietly falling behind the other.
var (
	_ Dialect = Postgres{}
	_ Dialect = MySQL{}
)

func TestNames(t *testing.T) {
	if got := (Postgres{}).Name(); got != "postgres" {
		t.Errorf("got %q", got)
	}
	if got := (MySQL{}).Name(); got != "mysql" {
		t.Errorf("got %q", got)
	}
}

func TestFullTextConditionDiffersByEngine(t *testing.T) {
	cols := []string{"title", "content"}

	pg, pgArg := (Postgres{}).FullTextCondition(cols, "investasi")
	for _, want := range []string{"to_tsvector", "plainto_tsquery", "coalesce(title, '')", "?"} {
		if !strings.Contains(pg, want) {
			t.Errorf("postgres condition %q missing %q", pg, want)
		}
	}
	if pgArg != "investasi" {
		t.Errorf("postgres arg = %v", pgArg)
	}

	my, myArg := (MySQL{}).FullTextCondition(cols, "investasi")
	for _, want := range []string{"MATCH (title, content)", "AGAINST (?", "NATURAL LANGUAGE MODE"} {
		if !strings.Contains(my, want) {
			t.Errorf("mysql condition %q missing %q", my, want)
		}
	}
	if myArg != "investasi" {
		t.Errorf("mysql arg = %v", myArg)
	}

	if pg == my {
		t.Error("full-text search is one of the two things that cannot be shared")
	}
}

// coalesce is not cosmetic: a NULL column would otherwise null the whole
// document and the row would stop matching anything at all.
func TestPostgresFullTextGuardsAgainstNullColumns(t *testing.T) {
	cond, _ := (Postgres{}).FullTextCondition([]string{"title", "content"}, "x")
	if strings.Count(cond, "coalesce(") != 2 {
		t.Errorf("every column must be coalesced, got %q", cond)
	}
}
