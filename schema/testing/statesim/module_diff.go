package statesim

import (
	"fmt"

	"cosmossdk.io/schema/view"
)

// DiffModuleStates compares the module state of two objects that implement ModuleState and returns a string with a diff if they
// are different or the empty string if they are the same.
func DiffModuleStates(expected, actual view.ModuleState) string {
	res := ""

	if expected.NumObjectCollections() != actual.NumObjectCollections() {
		res += fmt.Sprintf("OBJECT COLLECTION COUNT ERROR: expected %d, got %d\n", expected.NumObjectCollections(), actual.NumObjectCollections())
	}

	expected.ObjectCollections(func(expectedColl view.ObjectCollection) bool {
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
