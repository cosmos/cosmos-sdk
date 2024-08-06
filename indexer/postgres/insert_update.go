package postgres

import (
	"context"
	"fmt"
	"io"
	"strings"
)

// InsertUpdate inserts or updates the row with the provided key and value.
func (tm *ObjectIndexer) InsertUpdate(ctx context.Context, conn DBConn, key, value interface{}) error {
	exists, err := tm.Exists(ctx, conn, key)
	if err != nil {
		return err
	}

	buf := new(strings.Builder)
	var params []interface{}
	if exists {
		params, err = tm.UpdateSql(buf, key, value)
	} else {
		params, err = tm.InsertSql(buf, key, value)
	}

	sqlStr := buf.String()
	tm.options.Logger("Insert or Update", "sql", sqlStr, "params", params)
	_, err = conn.ExecContext(ctx, sqlStr, params...)
	return err
}

// InsertSql generates an INSERT statement and binding parameters for the provided key and value.
func (tm *ObjectIndexer) InsertSql(w io.Writer, key, value interface{}) ([]interface{}, error) {
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

	_, err = fmt.Fprintf(w, "INSERT INTO %q (%s) VALUES (%s);", tm.TableName(),
		strings.Join(allCols, ", "),
		strings.Join(paramBindings, ", "),
	)
	return allParams, err
}

// UpdateSql generates an UPDATE statement and binding parameters for the provided key and value.
func (tm *ObjectIndexer) UpdateSql(w io.Writer, key, value interface{}) ([]interface{}, error) {
	_, err := fmt.Fprintf(w, "UPDATE %q SET ", tm.TableName())

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

	if !tm.options.DisableRetainDeletions && tm.typ.RetainDeletions {
		_, err = fmt.Fprintf(w, ", _deleted = FALSE")
		if err != nil {
			return nil, err
		}
	}

	_, keyParams, err := tm.WhereSqlAndParams(w, key, paramIdx)
	if err != nil {
		return nil, err
	}

	allParams := append(valueParams, keyParams...)
	_, err = fmt.Fprintf(w, ";")
	return allParams, err
}
