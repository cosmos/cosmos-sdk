package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"strings"

	"cosmossdk.io/schema"
)

func (m *moduleManager) createEnumTypesForFields(ctx context.Context, tx *sql.Tx, fields []schema.Field) error {
	for _, field := range fields {
		if field.Kind != schema.EnumKind {
			continue
		}

		if _, ok := m.definedEnums[field.EnumDefinition.Name]; ok {
			// if the enum type is already defined, skip
			// we assume validation already happened
			continue
		}

		err := m.CreateEnumType(ctx, tx, field.EnumDefinition)
		if err != nil {
			return err
		}

		m.definedEnums[field.EnumDefinition.Name] = field.EnumDefinition
	}

	return nil
}

func enumTypeName(moduleName string, enum schema.EnumDefinition) string {
	return fmt.Sprintf("%s_%s", moduleName, enum.Name)
}

func (m *moduleManager) CreateEnumType(ctx context.Context, tx *sql.Tx, enum schema.EnumDefinition) error {
	typeName := enumTypeName(m.moduleName, enum)
	row := tx.QueryRowContext(ctx, "SELECT 1 FROM pg_type WHERE typname = $1", typeName)
	if err := row.Scan(); err != nil {
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

	// TODO: proper logging
	fmt.Printf("%s\n", buf.String())
	_, err = tx.ExecContext(ctx, buf.String())
	return err
}

func (m *moduleManager) CreateEnumTypeSql(writer io.Writer, enum schema.EnumDefinition) error {
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
