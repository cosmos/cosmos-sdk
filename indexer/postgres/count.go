package postgres

import (
	"context"
	"fmt"
)

// Count returns the number of rows in the table.
func (tm *ObjectIndexer) Count(ctx context.Context, conn DBConn) (int, error) {
	row := conn.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %q;", tm.TableName()))
	var count int
	err := row.Scan(&count)
	return count, err
}
