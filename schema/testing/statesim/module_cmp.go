package statesim

import (
	"io"

	"cosmossdk.io/schema"
)

type ModuleState interface {
	ModuleSchema() schema.ModuleSchema
	ObjectCollections(f func(value ObjectCollectionState) bool)
	GetObjectCollection(objectType string) (ObjectCollectionState, bool)
}

func CompareModuleStates(expected, actual ModuleState, diffWriter io.Writer) bool {
	panic("unimplemented")
}
