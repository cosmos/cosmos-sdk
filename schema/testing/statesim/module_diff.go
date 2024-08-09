package statesim

import (
	"fmt"

	"cosmossdk.io/schema"
)

// ModuleState defines an interface for things that represent module state in schema format.
type ModuleState interface {
	// ModuleSchema returns the schema for the module.
	ModuleSchema() schema.ModuleSchema

	// GetObjectCollection returns the object collection state for the given object type.
	GetObjectCollection(objectType string) (ObjectCollectionState, bool)

	// ObjectCollections iterates over all the object collection states in the module.
	ObjectCollections(f func(value ObjectCollectionState) bool)

	// NumObjectCollections returns the number of object collections in the module.
	NumObjectCollections() int
}

// DiffModuleStates compares the module state of two objects that implement ModuleState and returns a string with a diff if they
// are different or the empty string if they are the same.
func DiffModuleStates(expected, actual ModuleState) string {
	res := ""

	if expected.NumObjectCollections() != actual.NumObjectCollections() {
		res += fmt.Sprintf("OBJECT COLLECTION COUNT ERROR: expected %d, got %d\n", expected.NumObjectCollections(), actual.NumObjectCollections())
	}

	expected.ObjectCollections(func(expectedColl ObjectCollectionState) bool {
		objTypeName := expectedColl.ObjectType().Name
		actualColl, found := actual.GetObjectCollection(objTypeName)
		if !found {
			res += fmt.Sprintf("Object Collection %s: NOT FOUND\n", objTypeName)
			return true
		}

		diff := DiffObjectCollections(expectedColl, actualColl)
		if diff != "" {
			res += "Object Collection " + objTypeName + "\n"
			res += indentAllLines(diff)
		}

		return true
	})

	return res
}
