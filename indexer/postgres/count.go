package postgres

import (
	"context"
	"database/sql"
	"fmt"
)

func (tm *TableManager) Count(ctx context.Context, tx *sql.Tx) (int, error) {
	row := tx.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %q;", tm.TableName()))
	var count int
	err := row.Scan(&count)
	return count, err
}
