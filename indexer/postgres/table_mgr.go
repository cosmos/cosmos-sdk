package postgres

import (
	"fmt"

	"cosmossdk.io/schema"
)

// TableManager is a helper struct that generates SQL for a given object type.
type TableManager struct {
	moduleName  string
	typ         schema.ObjectType
	valueFields map[string]schema.Field
	allFields   map[string]schema.Field
	options     Options
}

// NewTableManager creates a new TableManager for the given object type.
func NewTableManager(moduleName string, typ schema.ObjectType, options Options) *TableManager {
	allFields := make(map[string]schema.Field)
	valueFields := make(map[string]schema.Field)

	for _, field := range typ.KeyFields {
		allFields[field.Name] = field
	}

	for _, field := range typ.ValueFields {
		valueFields[field.Name] = field
		allFields[field.Name] = field
	}

	return &TableManager{
		moduleName:  moduleName,
		typ:         typ,
		allFields:   allFields,
		valueFields: valueFields,
		options:     options,
	}
}

func (tm *TableManager) TableName() string {
	return fmt.Sprintf("%s_%s", tm.moduleName, tm.typ.Name)
}
