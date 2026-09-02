package dialect

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// MySQL is maintained and verified from an empty database, but is not the
// deployment target.
type MySQL struct{}

func (MySQL) Name() string { return "mysql" }

func (MySQL) FullTextCondition(columns []string, query string) (string, any) {
	return fmt.Sprintf(
		"MATCH (%s) AGAINST (? IN NATURAL LANGUAGE MODE)",
		strings.Join(columns, ", "),
	), query
}

// ResetSequence uses ALTER TABLE ... AUTO_INCREMENT. Neither the identifier nor
// the value can be bound as a parameter, so they are formatted in — safe here
// because the value is an int64 and the table name comes from application code,
// never from a request.
func (MySQL) ResetSequence(ctx context.Context, db *gorm.DB, table, _ string, next int64) error {
	stmt := fmt.Sprintf("ALTER TABLE `%s` AUTO_INCREMENT = %d", table, next)
	if err := db.WithContext(ctx).Exec(stmt).Error; err != nil {
		return fmt.Errorf("reset auto_increment for %s: %w", table, err)
	}
	return nil
}
