package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"strings"

	"cosmossdk.io/schema"
)

// createEnumType creates an enum type in the database.
func (m *moduleIndexer) createEnumType(ctx context.Context, conn dbConn, enum schema.EnumType) error {
	typeName := enumTypeName(m.moduleName, enum.Name)
	row := conn.QueryRowContext(ctx, "SELECT 1 FROM pg_type WHERE typname = $1", typeName)
	var res interface{}
	if err := row.Scan(&res); err != nil {
		if err != sql.ErrNoRows {
			return fmt.Errorf("failed to check if enum type %q exists: %v", typeName, err) //nolint:errorlint // using %v for go 1.12 compat
		}
	} else {
		// the enum type already exists
		// TODO: add a check to ensure the existing enum type matches the expected values, and update it if necessary?
		return nil
	}

	buf := new(strings.Builder)
	err := createEnumTypeSql(buf, m.moduleName, enum)
	if err != nil {
		return err
	}

	sqlStr := buf.String()
	if m.options.logger != nil {
		m.options.logger.Debug("Creating enum type", "sql", sqlStr)
	}
	_, err = conn.ExecContext(ctx, sqlStr)
	return err
}

// createEnumTypeSql generates a CREATE TYPE statement for the enum definition.
func createEnumTypeSql(writer io.Writer, moduleName string, enum schema.EnumType) error {
	_, err := fmt.Fprintf(writer, "CREATE TYPE %q AS ENUM (", enumTypeName(moduleName, enum.Name))
	if err != nil {
		return err
	}

	for i, value := range enum.Values {
		if i > 0 {
			_, err = fmt.Fprintf(writer, ", ")
			if err != nil {
				return err
			}
		}
		_, err = fmt.Fprintf(writer, "'%s'", value.Name)
		if err != nil {
			return err
		}
	}

	_, err = fmt.Fprintf(writer, ");")
	return err
}

// enumTypeName returns the name of the enum type scoped to the module.
func enumTypeName(moduleName, enumName string) string {
	return fmt.Sprintf("%s_%s", moduleName, enumName)
}
