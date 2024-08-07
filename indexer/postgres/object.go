package postgres

import (
	"fmt"

	"cosmossdk.io/schema"
)

// objectIndexer is a helper struct that generates SQL for a given object type.
type objectIndexer struct {
	moduleName  string
	typ         schema.ObjectType
	valueFields map[string]schema.Field
	allFields   map[string]schema.Field
	options     Options
}

// newObjectIndexer creates a new objectIndexer for the given object type.
func newObjectIndexer(moduleName string, typ schema.ObjectType, options Options) *objectIndexer {
	allFields := make(map[string]schema.Field)
	valueFields := make(map[string]schema.Field)

	for _, field := range typ.KeyFields {
		allFields[field.Name] = field
	}

	for _, field := range typ.ValueFields {
		valueFields[field.Name] = field
		allFields[field.Name] = field
	}

	return &objectIndexer{
		moduleName:  moduleName,
		typ:         typ,
		allFields:   allFields,
		valueFields: valueFields,
		options:     options,
	}
}

// TableName returns the name of the table for the object type scoped to its module.
func (tm *objectIndexer) TableName() string {
	return fmt.Sprintf("%s_%s", tm.moduleName, tm.typ.Name)
}
