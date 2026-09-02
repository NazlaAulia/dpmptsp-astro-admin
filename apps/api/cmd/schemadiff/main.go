// Command schemadiff proves the Postgres and MySQL migrations produce the same
// schema.
//
// Two migration sets are maintained because no single SQL file is valid on both
// engines. The obvious failure mode is that they drift: someone edits one, the
// other keeps working, and the difference is only discovered when the second
// engine is actually used. This turns that into a command that fails.
//
// It compares table names, column names, nullability and primary keys, with
// types normalised through an equivalence table — int4 and int are the same
// thing, and so are bool and tinyint(1).
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type column struct {
	Name     string
	Type     string
	Nullable bool
}

type table struct {
	Columns    map[string]column
	PrimaryKey []string
}

type schema map[string]table

func main() {
	var (
		pgDSN = flag.String("postgres", os.Getenv("SCHEMADIFF_POSTGRES"), "postgres DSN")
		myDSN = flag.String("mysql", os.Getenv("SCHEMADIFF_MYSQL"), "mysql DSN")
		myDB  = flag.String("mysql-database", "", "mysql database name (defaults to the DSN's)")
	)
	flag.Parse()

	if *pgDSN == "" || *myDSN == "" {
		fmt.Fprintln(os.Stderr, "both -postgres and -mysql DSNs are required")
		os.Exit(2)
	}

	pg, err := readPostgres(*pgDSN)
	if err != nil {
		fmt.Fprintf(os.Stderr, "postgres: %v\n", err)
		os.Exit(1)
	}
	my, err := readMySQL(*myDSN, *myDB)
	if err != nil {
		fmt.Fprintf(os.Stderr, "mysql: %v\n", err)
		os.Exit(1)
	}

	diffs := compare(pg, my)
	if len(diffs) == 0 {
		fmt.Printf("schemas agree: %d tables\n", len(pg))
		return
	}

	fmt.Printf("schemas differ (%d findings)\n\n", len(diffs))
	for _, d := range diffs {
		fmt.Println("  " + d)
	}
	os.Exit(1)
}

func compare(pg, my schema) []string {
	var out []string

	for _, name := range union(keys(pg), keys(my)) {
		p, inPG := pg[name]
		m, inMy := my[name]

		switch {
		case !inMy:
			out = append(out, fmt.Sprintf("table %-28s postgres only", name))
			continue
		case !inPG:
			out = append(out, fmt.Sprintf("table %-28s mysql only", name))
			continue
		}

		for _, col := range union(colNames(p), colNames(m)) {
			pc, okP := p.Columns[col]
			mc, okM := m.Columns[col]
			switch {
			case !okM:
				out = append(out, fmt.Sprintf("%s.%s postgres only", name, col))
			case !okP:
				out = append(out, fmt.Sprintf("%s.%s mysql only", name, col))
			default:
				if pc.Type != mc.Type {
					out = append(out, fmt.Sprintf("%s.%s type: postgres %s, mysql %s",
						name, col, pc.Type, mc.Type))
				}
				if pc.Nullable != mc.Nullable {
					out = append(out, fmt.Sprintf("%s.%s nullability: postgres nullable=%t, mysql nullable=%t",
						name, col, pc.Nullable, mc.Nullable))
				}
			}
		}

		if strings.Join(p.PrimaryKey, ",") != strings.Join(m.PrimaryKey, ",") {
			out = append(out, fmt.Sprintf("%s primary key: postgres [%s], mysql [%s]",
				name, strings.Join(p.PrimaryKey, ","), strings.Join(m.PrimaryKey, ",")))
		}
	}
	return out
}

// normalise maps each engine's type names onto a common vocabulary. Without
// this every column would be reported as different, and the tool would be
// useless.
func normalise(t string) string {
	t = strings.ToLower(strings.TrimSpace(t))
	switch t {
	case "int4", "integer", "int", "int unsigned", "mediumint":
		return "int"
	case "int8", "bigint":
		return "bigint"
	case "int2", "smallint":
		return "smallint"
	case "bool", "boolean", "tinyint", "tinyint(1)":
		return "bool"
	case "varchar", "character varying":
		return "varchar"
	case "text", "longtext", "mediumtext", "tinytext":
		return "text"
	case "timestamptz", "timestamp with time zone", "timestamp", "datetime":
		return "timestamp"
	case "date":
		return "date"
	case "numeric", "decimal":
		return "numeric"
	default:
		return t
	}
}

func readPostgres(dsn string) (schema, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	s := schema{}
	rows, err := db.Query(`
		SELECT table_name, column_name, data_type, is_nullable
		  FROM information_schema.columns
		 WHERE table_schema = 'public' AND table_name <> 'schema_migrations'
		 ORDER BY table_name, column_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var t, c, ty, nullable string
		if err := rows.Scan(&t, &c, &ty, &nullable); err != nil {
			return nil, err
		}
		addColumn(s, t, column{Name: c, Type: normalise(ty), Nullable: nullable == "YES"})
	}

	pk, err := db.Query(`
		SELECT tc.table_name, kcu.column_name
		  FROM information_schema.table_constraints tc
		  JOIN information_schema.key_column_usage kcu
		    ON kcu.constraint_name = tc.constraint_name
		 WHERE tc.constraint_type = 'PRIMARY KEY'
		   AND tc.table_schema = 'public'
		   AND tc.table_name <> 'schema_migrations'
		 ORDER BY tc.table_name, kcu.ordinal_position`)
	if err != nil {
		return nil, err
	}
	defer pk.Close()
	for pk.Next() {
		var t, c string
		if err := pk.Scan(&t, &c); err != nil {
			return nil, err
		}
		addPrimaryKey(s, t, c)
	}
	return s, rows.Err()
}

func readMySQL(dsn, dbName string) (schema, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	if dbName == "" {
		if err := db.QueryRow("SELECT DATABASE()").Scan(&dbName); err != nil {
			return nil, fmt.Errorf("resolve database name: %w", err)
		}
	}

	s := schema{}
	rows, err := db.Query(`
		SELECT table_name, column_name, data_type, is_nullable, column_key
		  FROM information_schema.columns
		 WHERE table_schema = ? AND table_name <> 'schema_migrations'
		 ORDER BY table_name, column_name`, dbName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var t, c, ty, nullable, key string
		if err := rows.Scan(&t, &c, &ty, &nullable, &key); err != nil {
			return nil, err
		}
		addColumn(s, t, column{Name: c, Type: normalise(ty), Nullable: nullable == "YES"})
		if key == "PRI" {
			addPrimaryKey(s, t, c)
		}
	}
	return s, rows.Err()
}

// ensure returns the entry for name, creating it if absent. It deliberately
// does not hand back a pointer into the map — Go does not allow that — so
// mutations go through addColumn and addPrimaryKey rather than through a copy
// that has to be written back by hand.
func ensure(s schema, name string) table {
	if _, ok := s[name]; !ok {
		s[name] = table{Columns: map[string]column{}}
	}
	return s[name]
}

func addColumn(s schema, tbl string, c column) {
	e := ensure(s, tbl)
	e.Columns[c.Name] = c
	s[tbl] = e
}

func addPrimaryKey(s schema, tbl, col string) {
	e := ensure(s, tbl)
	e.PrimaryKey = append(e.PrimaryKey, col)
	s[tbl] = e
}

func keys(s schema) []string {
	out := make([]string, 0, len(s))
	for k := range s {
		out = append(out, k)
	}
	return out
}

func colNames(t table) []string {
	out := make([]string, 0, len(t.Columns))
	for k := range t.Columns {
		out = append(out, k)
	}
	return out
}

func union(a, b []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range append(append([]string{}, a...), b...) {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	sort.Strings(out)
	return out
}
