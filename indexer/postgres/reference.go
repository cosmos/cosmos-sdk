package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"cosmossdk.io/schema"
)

func (m *moduleManager) createReferences(ctx context.Context, tx *sql.Tx) error {
	for _, objType := range m.schema.ObjectTypes {
		err := m.createReferencesForFields(ctx, tx, objType.KeyFields)
		if err != nil {
			return err
		}

		err = m.createReferencesForFields(ctx, tx, objType.ValueFields)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *moduleManager) createReferencesForFields(ctx context.Context, tx *sql.Tx, fields []schema.Field) error {
	for _, field := range fields {
		if field.References == "" {
			continue
		}

		err := m.createReference(ctx, tx, field)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *moduleManager) createReference(ctx context.Context, tx *sql.Tx, field schema.Field) error {
	parts := strings.Split(field.References, ".")
	if len(parts) != 2 {
		return fmt.Errorf("invalid reference: %s", field.References)
	}

	refTypeName, refFieldName := parts[0], parts[1]
	tbl, ok := m.tables[refTypeName]
	if !ok {
		return fmt.Errorf("unknown reference object type: %s", refTypeName)
	}

	refTypeName = tbl.TableName()

	fmt.Sprintf("ALTER TABLE %q ADD CONSTRAINT %q FOREIGN KEY (%q) REFERENCES %q (%q);")
}
