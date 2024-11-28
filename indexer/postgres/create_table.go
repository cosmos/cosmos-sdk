package postgres

import (
	"context"
	"fmt"
	"io"
	"strings"
)

// createTable creates the table for the object type.
func (tm *objectIndexer) createTable(ctx context.Context, conn dbConn) error {
	buf := new(strings.Builder)
	err := tm.createTableSql(buf)
	if err != nil {
		return err
	}

	sqlStr := buf.String()
	if tm.options.logger != nil {
		tm.options.logger.Debug("Creating table", "table", tm.tableName(), "sql", sqlStr)
	}
	_, err = conn.ExecContext(ctx, sqlStr)
	return err
}

// createTableSql generates a CREATE TABLE statement for the object type.
func (tm *objectIndexer) createTableSql(writer io.Writer) error {
	_, err := fmt.Fprintf(writer, "CREATE TABLE IF NOT EXISTS %q (\n\t", tm.tableName())
	if err != nil {
		return err
	}
	isSingleton := false
	if len(tm.typ.KeyFields) == 0 {
		isSingleton = true
		_, err = fmt.Fprintf(writer, "_id INTEGER NOT NULL CHECK (_id = 1),\n\t")
		if err != nil {
			return err
		}
	} else {
		for _, field := range tm.typ.KeyFields {
			err = tm.createColumnDefinition(writer, field)
			if err != nil {
				return err
			}
		}
	}

	for _, field := range tm.typ.ValueFields {
		err = tm.createColumnDefinition(writer, field)
		if err != nil {
			return err
		}
	}

	// add _deleted column when we have RetainDeletions set and enabled
	if !tm.options.disableRetainDeletions && tm.typ.RetainDeletions {
		_, err = fmt.Fprintf(writer, "_deleted BOOLEAN NOT NULL DEFAULT FALSE,\n\t")
		if err != nil {
			return err
		}
	}

	var pKeys []string
	if !isSingleton {
		for _, field := range tm.typ.KeyFields {
			name, err := tm.updatableColumnName(field)
			if err != nil {
				return err
			}

			pKeys = append(pKeys, name)
		}
	} else {
		pKeys = []string{"_id"}
	}

	_, err = fmt.Fprintf(writer, "PRIMARY KEY (%s)", strings.Join(pKeys, ", "))
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(writer, "\n);\n")
	if err != nil {
		return err
	}

	// we GRANT SELECT on the table to PUBLIC so that the table is automatically available
	// for querying using off-the-shelf tools like pg_graphql, Postgrest, Postgraphile, etc.
	// without any login permissions
	_, err = fmt.Fprintf(writer, "GRANT SELECT ON TABLE %q TO PUBLIC;", tm.tableName())
	if err != nil {
		return err
	}

	return nil
}
