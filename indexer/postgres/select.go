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
	params, err := tm.ExistsSqlAndParams(buf, key)
	if err != nil {
		return false, err
	}

	return tm.checkExists(ctx, tx, buf.String(), params)
}

func (tm *TableManager) ExistsSqlAndParams(w io.Writer, key interface{}) ([]interface{}, error) {
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

func (tm *TableManager) Equals(ctx context.Context, tx *sql.Tx, key, val interface{}) (bool, error) {
	buf := new(strings.Builder)
	params, err := tm.EqualsSqlAndParams(buf, key, val)
	if err != nil {
		return false, err
	}

	return tm.checkExists(ctx, tx, buf.String(), params)
}

func (tm *TableManager) EqualsSqlAndParams(w io.Writer, key, val interface{}) ([]interface{}, error) {
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

func (tm *TableManager) checkExists(ctx context.Context, tx *sql.Tx, sqlStr string, params []interface{}) (bool, error) {
	tm.options.Logger.Error("Select", "sql", sqlStr, "params", params)
	var res interface{}
	err := tx.QueryRowContext(ctx, sqlStr, params...).Scan(&res)
	switch err {
	case nil:
		return true, nil
	case sql.ErrNoRows:
		return false, nil
	default:
		return false, err
	}
}
