package postgres

import (
	"fmt"

	indexerbase "cosmossdk.io/indexer/base"
)

// TableManager is a helper struct that generates SQL for a given object type.
type TableManager struct {
	moduleName string
	typ        indexerbase.ObjectType
}

// NewTableManager creates a new TableManager for the given object type.
func NewTableManager(moduleName string, typ indexerbase.ObjectType) *TableManager {
	return &TableManager{
		moduleName: moduleName,
		typ:        typ,
	}
}

func (tm *TableManager) TableName() string {
	return fmt.Sprintf("%s_%s", tm.moduleName, tm.typ.Name)
}
