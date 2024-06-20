package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"strings"

	indexerbase "cosmossdk.io/indexer/base"
)

func (tm *TableManager) InsertUpdate(ctx context.Context, tx *sql.Tx, key, value interface{}) error {
	buf := new(strings.Builder)
	params, err := tm.InsertUpdateSqlAndParams(buf, key, value)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, buf.String(), params...)
	return err
}

func (tm *TableManager) InsertUpdateSqlAndParams(w io.Writer, key, value interface{}) ([]interface{}, error) {
	keyParams, err := tm.bindKeyParams(key)
	if err != nil {
		return nil, err
	}

	valueParams, valueCols, err := tm.bindValueParams(value)
	if err != nil {
		return nil, err
	}

	var keyCols []string
	if len(tm.typ.KeyFields) == 0 {
		keyCols = append(keyCols, "_id")
	} else {
		keyCols = colNames(tm.typ.KeyFields)
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
	return allParams, nil
}

func (tm *TableManager) bindKeyParams(key interface{}) ([]interface{}, error) {
	n := len(tm.typ.KeyFields)
	if n == 0 {
		// singleton, set _id = 1
		return []interface{}{1}, nil
	} else if n == 1 {
		return []interface{}{key}, nil
	} else {
		key, ok := key.([]interface{})
		if !ok {
			return nil, fmt.Errorf("expected key to be a slice")
		}

		return key, nil
	}
}

func (tm *TableManager) bindValueParams(value interface{}) (params []interface{}, valueCols []string, err error) {
	n := len(tm.typ.ValueFields)
	if n == 0 {
		return nil, nil, nil
	} else if valueUpdates, ok := value.(indexerbase.ValueUpdates); ok {
		var e error
		var fields []indexerbase.Field
		var params []interface{}
		if err := valueUpdates.Iterate(func(name string, value interface{}) bool {
			field, ok := tm.valueFields[name]
			if !ok {
				e = fmt.Errorf("unknown column %q", name)
				return false
			}
			fields = append(fields, field)
			params = append(params, value)
			return true
		}); err != nil {
			return nil, nil, err
		}
		if e != nil {
			return nil, nil, e
		}

		return params, colNames(fields), nil
	} else if n == 1 {
		return []interface{}{value}, []string{tm.typ.ValueFields[0].Name}, nil
	} else {
		value, ok := value.([]interface{})
		if !ok {
			return nil, nil, fmt.Errorf("expected value to be a slice")
		}

		return value, colNames(tm.typ.ValueFields), nil
	}
}
