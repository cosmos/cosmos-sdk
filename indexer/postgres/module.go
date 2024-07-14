package postgres

import (
	"context"
	"fmt"

	"cosmossdk.io/schema"
)

// ModuleIndexer manages the tables for a module.
type ModuleIndexer struct {
	moduleName   string
	schema       schema.ModuleSchema
	tables       map[string]*ObjectIndexer
	definedEnums map[string]schema.EnumDefinition
	options      Options
}

// NewModuleIndexer creates a new ModuleIndexer for the given module schema.
func NewModuleIndexer(moduleName string, modSchema schema.ModuleSchema, options Options) *ModuleIndexer {
	return &ModuleIndexer{
		moduleName:   moduleName,
		schema:       modSchema,
		tables:       map[string]*ObjectIndexer{},
		definedEnums: map[string]schema.EnumDefinition{},
		options:      options,
	}
}

// InitializeSchema creates tables for all object types in the module schema and creates enum types.
func (m *ModuleIndexer) InitializeSchema(ctx context.Context, conn DBConn) error {
	// create enum types
	for _, typ := range m.schema.ObjectTypes {
		err := m.createEnumTypesForFields(ctx, conn, typ.KeyFields)
		if err != nil {
			return err
		}

		err = m.createEnumTypesForFields(ctx, conn, typ.ValueFields)
		if err != nil {
			return err
		}
	}

	// create tables for all object types
	for _, typ := range m.schema.ObjectTypes {
		tm := NewObjectIndexer(m.moduleName, typ, m.options)
		m.tables[typ.Name] = tm
		err := tm.CreateTable(ctx, conn)
		if err != nil {
			return fmt.Errorf("failed to create table for %s in module %s: %v", typ.Name, m.moduleName, err) //nolint:errorlint // using %v for go 1.12 compat
		}
	}

	return nil
}

// ObjectIndexers returns the object indexers for the module.
func (m *ModuleIndexer) ObjectIndexers() map[string]*ObjectIndexer {
	return m.tables
}
