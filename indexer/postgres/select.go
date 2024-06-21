package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"strings"
)

func (tm *TableManager) Exists(ctx context.Context, tx *sql.Tx, key interface{}) (bool, error) {
	buf := new(strings.Builder)
	params, err := tm.SelectExistsSqlAndParams(buf, key)
	if err != nil {
		return false, err
	}

	// TODO: proper logging
	fmt.Printf("%s %v\n", buf.String(), params)
	var res interface{}
	err = tx.QueryRowContext(ctx, buf.String(), params...).Scan(&res)
	switch err {
	case nil:
		return true, nil
	case sql.ErrNoRows:
		return false, nil
	default:
		return false, err
	}
}

func (tm *TableManager) SelectExistsSqlAndParams(w io.Writer, key interface{}) ([]interface{}, error) {
	_, err := fmt.Fprintf(w, "SELECT 1 FROM %q", tm.TableName())
	if err != nil {
		return nil, err
	}

	_, keyParams, err := tm.WhereSqlAndParams(w, key, 1)
	if err != nil {
		return nil, err
	}

	_, err = fmt.Fprintf(w, ";")
	return keyParams, err
}
