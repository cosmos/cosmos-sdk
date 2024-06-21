package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"strings"
)

func (tm *TableManager) InsertUpdate(ctx context.Context, tx *sql.Tx, key, value interface{}) error {
	buf := new(strings.Builder)
	params, err := tm.InsertUpdateSqlAndParams(buf, key, value)
	if err != nil {
		return err
	}

	// TODO: proper logging
	fmt.Printf("%s %v\n", buf.String(), params)
	_, err = tx.ExecContext(ctx, buf.String(), params...)
	return err
}

func (tm *TableManager) InsertUpdateSqlAndParams(w io.Writer, key, value interface{}) ([]interface{}, error) {
	keyParams, keyCols, err := tm.bindKeyParams(key)
	if err != nil {
		return nil, err
	}

	valueParams, valueCols, err := tm.bindValueParams(value)
	if err != nil {
		return nil, err
	}

	allCols := make([]string, 0, len(keyCols)+len(valueCols))
	allCols = append(allCols, keyCols...)
	allCols = append(allCols, valueCols...)

	var paramBindings []string
	for i := 1; i <= len(allCols); i++ {
		paramBindings = append(paramBindings, fmt.Sprintf("$%d", i))
	}

	_, err = fmt.Fprintf(w, "INSERT INTO %q (%s) VALUES (%s) ON CONFLICT (%s) DO ", tm.TableName(),
		strings.Join(allCols, ", "),
		strings.Join(paramBindings, ", "),
		strings.Join(keyCols, ", "),
	)
	if err != nil {
		return nil, err
	}

	if len(valueCols) == 0 {
		_, err = fmt.Fprintf(w, "NOTHING")
		if err != nil {
			return nil, err
		}
		return keyParams, nil
	}

	_, err = fmt.Fprintf(w, "UPDATE SET ")
	if err != nil {
		return nil, err
	}

	for i, col := range valueCols {
		if i > 0 {
			_, err = fmt.Fprintf(w, ", ")
			if err != nil {
				return nil, err
			}
		}
		_, err = fmt.Fprintf(w, "%s = EXCLUDED.%s", col, col)
		if err != nil {
			return nil, err
		}
	}

	var allParams []interface{}
	allParams = append(allParams, keyParams...)
	allParams = append(allParams, valueParams...)

	_, err = fmt.Fprintf(w, ";")
	return allParams, err
}
