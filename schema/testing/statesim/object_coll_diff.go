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
	if expected.Len() != actual.Len() {
		res += fmt.Sprintf("OBJECT COUNT ERROR: expected %d, got %d\n", expected.Len(), actual.Len())
	}

	expected.AllState(func(expectedUpdate schema.ObjectUpdate) bool {
		actualUpdate, found := actual.GetObject(expectedUpdate.Key)
		if !found {
			res += fmt.Sprintf("Object %s: NOT FOUND\n", fmtObjectKey(expected.ObjectType(), expectedUpdate.Key))
			return true
		}

		if expectedUpdate.Delete != actualUpdate.Delete {
			res += fmt.Sprintf("Object %s: Deleted mismatch, expected %v, got %v\n", fmtObjectKey(expected.ObjectType(), expectedUpdate.Key), expectedUpdate.Delete, actualUpdate.Delete)
		}

		if expectedUpdate.Delete {
			return true
		}

		valueDiff := schematesting.DiffObjectValues(expected.ObjectType().ValueFields, expectedUpdate.Value, actualUpdate.Value)
		if valueDiff != "" {
			res += "Object "
			res += fmtObjectKey(expected.ObjectType(), expectedUpdate.Key)
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
