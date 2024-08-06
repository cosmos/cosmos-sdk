package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"strings"
)

// Exists checks if a row with the provided key exists in the table.
func (tm *ObjectIndexer) Exists(ctx context.Context, conn DBConn, key interface{}) (bool, error) {
	buf := new(strings.Builder)
	params, err := tm.ExistsSqlAndParams(buf, key)
	if err != nil {
		return false, err
	}

	return tm.checkExists(ctx, conn, buf.String(), params)
}

// ExistsSqlAndParams generates a SELECT statement to check if a row with the provided key exists in the table.
func (tm *ObjectIndexer) ExistsSqlAndParams(w io.Writer, key interface{}) ([]interface{}, error) {
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

// Equals checks if a row with the provided key and value exists.
func (tm *ObjectIndexer) Equals(ctx context.Context, conn DBConn, key, val interface{}) (bool, error) {
	buf := new(strings.Builder)
	params, err := tm.EqualsSqlAndParams(buf, key, val)
	if err != nil {
		return false, err
	}

	return tm.checkExists(ctx, conn, buf.String(), params)
}

// EqualsSqlAndParams generates a SELECT statement to check if a row with the provided key and value exists in the table.
func (tm *ObjectIndexer) EqualsSqlAndParams(w io.Writer, key, val interface{}) ([]interface{}, error) {
	_, err := fmt.Fprintf(w, "SELECT 1 FROM %q", tm.TableName())
	if err != nil {
		return nil, err
	}

	keyParams, keyCols, err := tm.bindKeyParams(key)
	if err != nil {
		return nil, err
	}

	valueParams, valueCols, err := tm.bindValueParams(val)
	if err != nil {
		return nil, err
	}

	allParams := make([]interface{}, 0, len(keyParams)+len(valueParams))
	allParams = append(allParams, keyParams...)
	allParams = append(allParams, valueParams...)

	allCols := make([]string, 0, len(keyCols)+len(valueCols))
	allCols = append(allCols, keyCols...)
	allCols = append(allCols, valueCols...)

	_, allParams, err = tm.WhereSql(w, allParams, allCols, 1)
	if err != nil {
		return nil, err
	}

	_, err = fmt.Fprintf(w, ";")
	return allParams, err
}

// checkExists checks if a row exists in the table.
func (tm *ObjectIndexer) checkExists(ctx context.Context, conn DBConn, sqlStr string, params []interface{}) (bool, error) {
	tm.options.Logger.Debug("Select", "sql", sqlStr, "params", params)
	var res interface{}
	// TODO check for multiple rows which would be a logic error
	err := conn.QueryRowContext(ctx, sqlStr, params...).Scan(&res)
	switch err {
	case nil:
		return true, nil
	case sql.ErrNoRows:
		return false, nil
	default:
		return false, err
	}
}
