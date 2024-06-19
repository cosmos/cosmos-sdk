package state

import (
	"github.com/tidwall/btree"
	"pgregory.net/rapid"

	indexerbase "cosmossdk.io/indexer/base"
)

type Entry struct {
	Key   any
	Value any
}

type AppState struct {
	Modules *btree.Map[string, *ModuleState]
}

type ModuleState struct {
	ModuleSchema indexerbase.ModuleSchema
	Objects      *btree.Map[string, *ObjectState]
}

type ObjectState struct {
	ObjectType indexerbase.ObjectType
	Objects    *btree.Map[string, *Entry]
	UpdateGen  *rapid.Generator[indexerbase.ObjectUpdate]
}
