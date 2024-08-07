package postgres

import (
	"context"
	"fmt"
	"io"
	"strings"
)

// delete deletes the row with the provided key from the table.
func (tm *objectIndexer) delete(ctx context.Context, conn dbConn, key interface{}) error {
	buf := new(strings.Builder)
	var params []interface{}
	var err error
	if !tm.options.DisableRetainDeletions && tm.typ.RetainDeletions {
		params, err = tm.RetainDeleteSqlAndParams(buf, key)
	} else {
		params, err = tm.deleteSqlAndParams(buf, key)
	}
	if err != nil {
		return err
	}

	sqlStr := buf.String()
	tm.options.Logger("Delete", "sql", sqlStr, "params", params)
	_, err = conn.ExecContext(ctx, sqlStr, params...)
	return err
}

// deleteSqlAndParams generates a DELETE statement and binding parameters for the provided key.
func (tm *objectIndexer) deleteSqlAndParams(w io.Writer, key interface{}) ([]interface{}, error) {
	_, err := fmt.Fprintf(w, "DELETE FROM %q", tm.TableName())
	if err != nil {
		return nil, err
	}

	_, keyParams, err := tm.whereSqlAndParams(w, key, 1)
	if err != nil {
		return nil, err
	}

	_, err = fmt.Fprintf(w, ";")
	return keyParams, err
}

// RetainDeleteSqlAndParams generates an UPDATE statement to set the _deleted column to true for the provided key
// which is used when the table is set to retain deletions mode.
func (tm *objectIndexer) RetainDeleteSqlAndParams(w io.Writer, key interface{}) ([]interface{}, error) {
	_, err := fmt.Fprintf(w, "UPDATE %q SET _deleted = TRUE", tm.TableName())
	if err != nil {
		return nil, err
	}

	_, keyParams, err := tm.whereSqlAndParams(w, key, 1)
	if err != nil {
		return nil, err
	}

	_, err = fmt.Fprintf(w, ";")
	return keyParams, err
}
