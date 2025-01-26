package statesim

import (
	"fmt"

	"cosmossdk.io/schema/view"
)

// DiffModuleStates compares the module state of two objects that implement ModuleState and returns a string with a diff if they
// are different or the empty string if they are the same.
func DiffModuleStates(expected, actual view.ModuleState) string {
	res := ""

	expectedNumObjectCollections, err := expected.NumObjectCollections()
	if err != nil {
		res += fmt.Sprintf("ERROR getting expected num object collections: %s\n", err)
		return res
	}

	actualNumObjectCollections, err := actual.NumObjectCollections()
	if err != nil {
		res += fmt.Sprintf("ERROR getting actual num object collections: %s\n", err)
		return res
	}

	if expectedNumObjectCollections != actualNumObjectCollections {
		res += fmt.Sprintf("OBJECT COLLECTION COUNT ERROR: expected %d, got %d\n", expectedNumObjectCollections, actualNumObjectCollections)
	}

	for expectedColl, err := range expected.ObjectCollections {
		if err != nil {
			res += fmt.Sprintf("ERROR getting expected object collection: %s\n", err)
			continue
		}

		objTypeName := expectedColl.ObjectType().Name
		actualColl, err := actual.GetObjectCollection(objTypeName)
		if err != nil {
			res += fmt.Sprintf("ERROR getting actual object collection: %s\n", err)
			continue
		}
		if actualColl == nil {
			res += fmt.Sprintf("Object Collection %s: actual collection NOT FOUND\n", objTypeName)
			continue
		}

		diff := DiffObjectCollections(expectedColl, actualColl)
		if diff != "" {
			res += "Object Collection " + objTypeName + "\n"
			res += indentAllLines(diff)
		}
	}

	return res
}
