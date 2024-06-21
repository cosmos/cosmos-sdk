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
	keyParams, keyCols, err := tm.bindKeyParams(key)
	if err != nil {
		return nil, err
	}

	_, err = fmt.Fprintf(w, "DELETE FROM %q WHERE ", tm.TableName())
	if err != nil {
		return nil, err
	}

	for i, col := range keyCols {
		if i > 0 {
			_, err = fmt.Fprintf(w, " AND ")
			if err != nil {
				return nil, err
			}
		}
		_, err = fmt.Fprintf(w, "%s = $%d", col, i)
		if err != nil {
			return nil, err
		}
	}

	_, err = fmt.Fprintf(w, ";")
	return keyParams, err
}
