package statesim

import (
	"fmt"
	"strings"

	"cosmossdk.io/schema"
	schematesting "cosmossdk.io/schema/testing"
	"cosmossdk.io/schema/view"
)

// DiffObjectCollections compares the object collection state of two objects that implement ObjectCollectionState and returns a string with a diff if they
// are different or the empty string if they are the same.
func DiffObjectCollections(expected, actual view.ObjectCollection) string {
	res := ""

	expectedNumObjects, err := expected.Len()
	if err != nil {
		res += fmt.Sprintf("ERROR getting expected num objects: %s\n", err)
		return res
	}

	actualNumObjects, err := actual.Len()
	if err != nil {
		res += fmt.Sprintf("ERROR getting actual num objects: %s\n", err)
		return res
	}

	if expectedNumObjects != actualNumObjects {
		res += fmt.Sprintf("OBJECT COUNT ERROR: expected %d, got %d\n", expectedNumObjects, actualNumObjects)
	}

	expected.AllState(func(expectedUpdate schema.ObjectUpdate, err error) bool {
		if err != nil {
			res += fmt.Sprintf("ERROR getting expected object: %s\n", err)
			return true
		}

		keyStr := fmtObjectKey(expected.ObjectType(), expectedUpdate.Key)
		actualUpdate, found, err := actual.GetObject(expectedUpdate.Key)
		if err != nil {
			res += fmt.Sprintf("Object %s: ERROR: %v\n", keyStr, err)
			return true
		}
		if !found {
			res += fmt.Sprintf("Object %s: NOT FOUND\n", keyStr)
			return true
		}

		if expectedUpdate.Delete != actualUpdate.Delete {
			res += fmt.Sprintf("Object %s: Deleted mismatch, expected %v, got %v\n", keyStr, expectedUpdate.Delete, actualUpdate.Delete)
		}

		if expectedUpdate.Delete {
			return true
		}

		valueDiff := schematesting.DiffObjectValues(expected.ObjectType().ValueFields, expectedUpdate.Value, actualUpdate.Value)
		if valueDiff != "" {
			res += "Object "
			res += keyStr
			res += "\n"
			res += indentAllLines(valueDiff)
		}

		return true
	})

	return res
}

func indentAllLines(s string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = "  " + line
	}
	return strings.Join(lines, "\n")
}
