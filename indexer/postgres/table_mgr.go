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
}

// NewTableManager creates a new TableManager for the given object type.
func NewTableManager(moduleName string, typ schema.ObjectType) *TableManager {
	valueFields := make(map[string]schema.Field)
	for _, field := range typ.ValueFields {
		valueFields[field.Name] = field
	}

	return &TableManager{
		moduleName: moduleName,
		typ:        typ,
	}
}

func (tm *TableManager) TableName() string {
	return fmt.Sprintf("%s_%s", tm.moduleName, tm.typ.Name)
}
