package postgres

import (
	"context"
	"fmt"

	"cosmossdk.io/schema"
)

// moduleIndexer manages the tables for a module.
type moduleIndexer struct {
	moduleName   string
	schema       schema.ModuleSchema
	tables       map[string]*objectIndexer
	definedEnums map[string]schema.EnumType
	options      options
}

// newModuleIndexer creates a new moduleIndexer for the given module schema.
func newModuleIndexer(moduleName string, modSchema schema.ModuleSchema, options options) *moduleIndexer {
	return &moduleIndexer{
		moduleName:   moduleName,
		schema:       modSchema,
		tables:       map[string]*objectIndexer{},
		definedEnums: map[string]schema.EnumType{},
		options:      options,
	}
}

// initializeSchema creates tables for all object types in the module schema and creates enum types.
func (m *moduleIndexer) initializeSchema(ctx context.Context, conn dbConn) error {
	// create enum types
	var err error
	m.schema.EnumTypes(func(enumType schema.EnumType) bool {
		err = m.createEnumType(ctx, conn, enumType)
		return err == nil
	})
	if err != nil {
		return err
	}

	// create tables for all object types
	m.schema.StateObjectTypes(func(typ schema.StateObjectType) bool {
		tm := newObjectIndexer(m.moduleName, typ, m.options)
		m.tables[typ.Name] = tm
		err = tm.createTable(ctx, conn)
		if err != nil {
			err = fmt.Errorf("failed to create table for %s in module %s: %v", typ.Name, m.moduleName, err) //nolint:errorlint // using %v for go 1.12 compat
		}
		return err == nil
	})

	return err
}
