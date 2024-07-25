package statesim

import (
	"fmt"
	"strings"

	"cosmossdk.io/schema"
	schematesting "cosmossdk.io/schema/testing"
)

type ObjectCollectionState interface {
	AllState(f func(schema.ObjectUpdate) bool)
	GetObject(key any) (update schema.ObjectUpdate, found bool)
	ObjectType() schema.ObjectType
	Len() int
}

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

		valueDiff := schematesting.DiffObjectValues(expected.ObjectType().KeyFields, expectedUpdate.Value, actualUpdate.Value)
		if valueDiff != "" {
			res += "Object "
			res += fmtObjectKey(expected.ObjectType(), expectedUpdate.Key)
			res += "\n"
			res += indentAllLines(valueDiff)
		}

		panic("check Deleted")

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
