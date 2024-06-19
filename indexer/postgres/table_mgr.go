package postgres

import (
	"fmt"

	indexerbase "cosmossdk.io/indexer/base"
)

// TableManager is a helper struct that generates SQL for a given object type.
type TableManager struct {
	moduleName  string
	typ         indexerbase.ObjectType
	valueFields map[string]indexerbase.Field
}

// NewTableManager creates a new TableManager for the given object type.
func NewTableManager(moduleName string, typ indexerbase.ObjectType) *TableManager {
	valueFields := make(map[string]indexerbase.Field)
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
