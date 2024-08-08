package statesim

import (
	"fmt"
	"strings"

	"cosmossdk.io/schema"
	schematesting "cosmossdk.io/schema/testing"
)

// ObjectCollectionState is the interface for the state of an object collection
// represented by ObjectUpdate's for an ObjectType. ObjectUpdates must not include
// ValueUpdates in the Value field. When ValueUpdates are applied they must be
// converted to individual value or array format depending on the number of fields in
// the value. For collections which retain deletions, ObjectUpdate's with the Remove
// field set to true should be returned with the latest Value still intact.
type ObjectCollectionState interface {
	// ObjectType returns the object type for the collection.
	ObjectType() schema.ObjectType

	// GetObject returns the object update for the given key if it exists.
	GetObject(key any) (update schema.ObjectUpdate, found bool)

	// AllState iterates over the state of the collection by calling the given function with each item in
	// state represented as an object update.
	AllState(f func(schema.ObjectUpdate) bool)

	// Len returns the number of objects in the collection.
	Len() int
}

// DiffObjectCollections compares the object collection state of two objects that implement ObjectCollectionState and returns a string with a diff if they
// are different or the empty string if they are the same.
func DiffObjectCollections(expected, actual ObjectCollectionState) string {
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
