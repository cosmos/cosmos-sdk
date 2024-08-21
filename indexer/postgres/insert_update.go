package postgres

import (
	"context"
	"fmt"
	"io"
	"strings"
)

// insertUpdate inserts or updates the row with the provided key and value.
func (tm *objectIndexer) insertUpdate(ctx context.Context, conn dbConn, key, value interface{}) error {
	exists, err := tm.exists(ctx, conn, key)
	if err != nil {
		return err
	}

	buf := new(strings.Builder)
	var params []interface{}
	if exists {
		if len(tm.typ.ValueFields) == 0 {
			// special case where there are no value fields, so we can't update anything
			return nil
		}

		params, err = tm.updateSql(buf, key, value)
	} else {
		params, err = tm.insertSql(buf, key, value)
	}
	if err != nil {
		return err
	}

	sqlStr := buf.String()
	if tm.options.logger != nil {
		tm.options.logger.Debug("Insert or Update", "sql", sqlStr, "params", params)
	}
	_, err = conn.ExecContext(ctx, sqlStr, params...)
	return err
}

// insertSql generates an INSERT statement and binding parameters for the provided key and value.
func (tm *objectIndexer) insertSql(w io.Writer, key, value interface{}) ([]interface{}, error) {
	keyParams, keyCols, err := tm.bindKeyParams(key)
	if err != nil {
		return nil, err
	}

	valueParams, valueCols, err := tm.bindValueParams(value)
	if err != nil {
		return nil, err
	}

	var allParams []interface{}
	allParams = append(allParams, keyParams...)
	allParams = append(allParams, valueParams...)

	allCols := make([]string, 0, len(keyCols)+len(valueCols))
	allCols = append(allCols, keyCols...)
	allCols = append(allCols, valueCols...)

	var paramBindings []string
	for i := 1; i <= len(allCols); i++ {
		paramBindings = append(paramBindings, fmt.Sprintf("$%d", i))
	}

	_, err = fmt.Fprintf(w, "INSERT INTO %q (%s) VALUES (%s);", tm.tableName(),
		strings.Join(allCols, ", "),
		strings.Join(paramBindings, ", "),
	)
	return allParams, err
}

// updateSql generates an UPDATE statement and binding parameters for the provided key and value.
func (tm *objectIndexer) updateSql(w io.Writer, key, value interface{}) ([]interface{}, error) {
	_, err := fmt.Fprintf(w, "UPDATE %q SET ", tm.tableName())
	if err != nil {
		return nil, err
	}

	valueParams, valueCols, err := tm.bindValueParams(value)
	if err != nil {
		return nil, err
	}

	paramIdx := 1
	for i, col := range valueCols {
		if i > 0 {
			_, err = fmt.Fprintf(w, ", ")
			if err != nil {
				return nil, err
			}
		}
		_, err = fmt.Fprintf(w, "%s = $%d", col, paramIdx)
		if err != nil {
			return nil, err
		}

		paramIdx++
	}

	if !tm.options.disableRetainDeletions && tm.typ.RetainDeletions {
		_, err = fmt.Fprintf(w, ", _deleted = FALSE")
		if err != nil {
			return nil, err
		}
	}

	_, keyParams, err := tm.whereSqlAndParams(w, key, paramIdx)
	if err != nil {
		return nil, err
	}

	allParams := append(valueParams, keyParams...)
	_, err = fmt.Fprintf(w, ";")
	return allParams, err
}
