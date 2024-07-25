package statesim

import (
	"fmt"

	"cosmossdk.io/schema"
)

type ModuleState interface {
	ModuleSchema() schema.ModuleSchema
	ObjectCollections(f func(value ObjectCollectionState) bool)
	GetObjectCollection(objectType string) (ObjectCollectionState, bool)
	NumObjectCollections() int
}

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
