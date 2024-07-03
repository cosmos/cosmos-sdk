package postgres

import (
	"context"
	"fmt"

	"cosmossdk.io/schema"
)

// ModuleManager manages the tables for a module.
type ModuleManager struct {
	moduleName   string
	schema       schema.ModuleSchema
	tables       map[string]*TableManager
	definedEnums map[string]schema.EnumDefinition
	options      Options
}

// NewModuleManager creates a new ModuleManager for the given module schema.
func NewModuleManager(moduleName string, modSchema schema.ModuleSchema, options Options) *ModuleManager {
	return &ModuleManager{
		moduleName:   moduleName,
		schema:       modSchema,
		tables:       map[string]*TableManager{},
		definedEnums: map[string]schema.EnumDefinition{},
		options:      options,
	}
}

// InitializeSchema creates tables for all object types in the module schema and creates enum types.
func (m *ModuleManager) InitializeSchema(ctx context.Context, conn DBConn) error {
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
		tm := NewTableManager(m.moduleName, typ, m.options)
		m.tables[typ.Name] = tm
		err := tm.CreateTable(ctx, conn)
		if err != nil {
			return fmt.Errorf("failed to create table for %s in module %s: %w", typ.Name, m.moduleName, err)
		}
	}

	return nil

}

// Tables returns the table managers for the module.
func (m *ModuleManager) Tables() map[string]*TableManager {
	return m.tables
}
