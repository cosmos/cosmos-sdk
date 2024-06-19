package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"strings"
)

func (tm *TableManager) CreateTable(ctx context.Context, tx *sql.Tx) error {
	buf := new(strings.Builder)
	err := tm.CreateTableSql(buf)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, buf.String())
	return err
}

// CreateTableSql generates a CREATE TABLE statement for the object type.
func (tm *TableManager) CreateTableSql(writer io.Writer) error {
	_, err := fmt.Fprintf(writer, "CREATE TABLE IF NOT EXISTS %q (", tm.TableName())
	if err != nil {
		return err
	}
	isSingleton := false
	if len(tm.typ.KeyFields) == 0 {
		isSingleton = true
		_, err = fmt.Fprintf(writer, "_id INTEGER NOT NULL CHECK (_id = 1),\n\t")
	} else {
		for _, field := range tm.typ.KeyFields {
			err = tm.createColumnDef(writer, field)
			if err != nil {
				return err
			}
		}
	}

	for _, field := range tm.typ.ValueFields {
		err = tm.createColumnDef(writer, field)
		if err != nil {
			return err
		}
	}

	var pKeys []string
	if !isSingleton {
		for _, field := range tm.typ.KeyFields {
			pKeys = append(pKeys, `"`+field.Name+`"`)
		}
	} else {
		pKeys = []string{"_id"}
	}

	_, err = fmt.Fprintf(writer, "PRIMARY KEY (%s)\n", strings.Join(pKeys, ", "))
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(writer, `);

GRANT SELECT ON TABLE %q TO PUBLIC;
`, tm.TableName())
	if err != nil {
		return err
	}

	return nil
}
