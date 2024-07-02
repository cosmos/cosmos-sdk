package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"strings"

	"cosmossdk.io/schema"
)

// createEnumTypesForFields creates enum types for all the fields that have enum kind in the module schema.
func (m *ModuleManager) createEnumTypesForFields(ctx context.Context, conn DBConn, fields []schema.Field) error {
	for _, field := range fields {
		if field.Kind != schema.EnumKind {
			continue
		}

		if _, ok := m.definedEnums[field.EnumDefinition.Name]; ok {
			// if the enum type is already defined, skip
			// we assume validation already happened
			continue
		}

		err := m.CreateEnumType(ctx, conn, field.EnumDefinition)
		if err != nil {
			return err
		}

		m.definedEnums[field.EnumDefinition.Name] = field.EnumDefinition
	}

	return nil
}

// CreateEnumType creates an enum type in the database.
func (m *ModuleManager) CreateEnumType(ctx context.Context, conn DBConn, enum schema.EnumDefinition) error {
	typeName := enumTypeName(m.moduleName, enum)
	row := conn.QueryRowContext(ctx, "SELECT 1 FROM pg_type WHERE typname = $1", typeName)
	var res interface{}
	if err := row.Scan(&res); err != nil {
		if err != sql.ErrNoRows {
			return fmt.Errorf("failed to check if enum type %q exists: %w", typeName, err)
		}
	} else {
		// the enum type already exists
		return nil
	}

	buf := new(strings.Builder)
	err := m.CreateEnumTypeSql(buf, enum)
	if err != nil {
		return err
	}

	sqlStr := buf.String()
	m.options.Logger.Debug("Creating enum type", "sql", sqlStr)
	_, err = conn.ExecContext(ctx, sqlStr)
	return err
}

// CreateEnumTypeSql generates a CREATE TYPE statement for the enum definition.
func (m *ModuleManager) CreateEnumTypeSql(writer io.Writer, enum schema.EnumDefinition) error {
	_, err := fmt.Fprintf(writer, "CREATE TYPE %q AS ENUM (", enumTypeName(m.moduleName, enum))
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
		_, err = fmt.Fprintf(writer, "'%s'", value)
		if err != nil {
			return err
		}
	}

	_, err = fmt.Fprintf(writer, ");")
	return err
}

// enumTypeName returns the name of the enum type scoped to the module.
func enumTypeName(moduleName string, enum schema.EnumDefinition) string {
	return fmt.Sprintf("%s_%s", moduleName, enum.Name)
}
