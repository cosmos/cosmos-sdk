package postgres

import (
	indexerbase "cosmossdk.io/indexer/base"
)

// SQLGenerator is a helper struct that generates SQL for a given object type.
type SQLGenerator struct {
	typ indexerbase.ObjectType
}

// NewSQLGenerator creates a new SQLGenerator for the given object type.
func NewSQLGenerator(typ indexerbase.ObjectType) *SQLGenerator {
	return &SQLGenerator{typ: typ}
}
