package view

import "cosmossdk.io/schema"

// ModuleState defines an interface for things that represent module state in schema format.
type ModuleState interface {
	// ModuleSchema returns the schema for the module.
	ModuleSchema() schema.ModuleSchema

	// GetObjectCollection returns the object collection for the given object type.
	GetObjectCollection(objectType string) (ObjectCollection, bool)

	// ObjectCollections iterates over all the object collections in the module.
	ObjectCollections(f func(value ObjectCollection) bool)

	// NumObjectCollections returns the number of object collections in the module.
	NumObjectCollections() int
}
