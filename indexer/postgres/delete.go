package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"strings"
)

func (tm *TableManager) Delete(ctx context.Context, tx *sql.Tx, key interface{}) error {
	buf := new(strings.Builder)
	params, err := tm.DeleteSqlAndParams(buf, key)
	if err != nil {
		return err
	}

	// TODO: proper logging
	fmt.Printf("%s %v\n", buf.String(), params)
	_, err = tx.ExecContext(ctx, buf.String(), params...)
	return err
}

func (tm *TableManager) DeleteSqlAndParams(w io.Writer, key interface{}) ([]interface{}, error) {
	_, err := fmt.Fprintf(w, "DELETE FROM %q", tm.TableName())
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
