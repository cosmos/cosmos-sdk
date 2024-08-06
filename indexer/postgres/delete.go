package postgres

import (
	"context"
	"fmt"
	"io"
	"strings"
)

// Delete deletes the row with the provided key from the table.
func (tm *ObjectIndexer) Delete(ctx context.Context, conn DBConn, key interface{}) error {
	buf := new(strings.Builder)
	var params []interface{}
	var err error
	if !tm.options.DisableRetainDeletions && tm.typ.RetainDeletions {
		params, err = tm.RetainDeleteSqlAndParams(buf, key)
	} else {
		params, err = tm.DeleteSqlAndParams(buf, key)
	}
	if err != nil {
		return err
	}

	sqlStr := buf.String()
	tm.options.Logger("Delete", "sql", sqlStr, "params", params)
	_, err = conn.ExecContext(ctx, sqlStr, params...)
	return err
}

// DeleteSqlAndParams generates a DELETE statement and binding parameters for the provided key.
func (tm *ObjectIndexer) DeleteSqlAndParams(w io.Writer, key interface{}) ([]interface{}, error) {
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

// RetainDeleteSqlAndParams generates an UPDATE statement to set the _deleted column to true for the provided key
// which is used when the table is set to retain deletions mode.
func (tm *ObjectIndexer) RetainDeleteSqlAndParams(w io.Writer, key interface{}) ([]interface{}, error) {
	_, err := fmt.Fprintf(w, "UPDATE %q SET _deleted = TRUE", tm.TableName())
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
