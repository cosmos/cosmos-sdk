package statesim

import "cosmossdk.io/schema"

type ModuleState interface {
	ModuleSchema() schema.ModuleSchema
	ObjectCollections(f func(value ObjectCollectionState) bool)
	GetObjectCollection(objectType string) (ObjectCollectionState, bool)
}
