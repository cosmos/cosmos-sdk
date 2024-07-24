package statesim

import (
	"fmt"
	"io"

	"cosmossdk.io/schema"
	schematesting "cosmossdk.io/schema/testing"
)

type ObjectCollectionState interface {
	AllState(f func(schema.ObjectUpdate) bool)
	GetObject(key any) (update schema.ObjectUpdate, found bool)
	ObjectType() schema.ObjectType
	Len() int
}

func CompareObjectCollections(expected, actual ObjectCollectionState, diffWriter io.Writer) bool {
	if expected.Len() != actual.Len() {
		_, err := io.WriteString(diffWriter, "ObjectCollection length mismatch")
		if err != nil {
			panic(err)
		}
		return false
	}

	eq := true

	expected.AllState(func(expectedUpdate schema.ObjectUpdate) bool {
		actualUpdate, found := actual.GetObject(expectedUpdate.Key)
		if !found {
			_, err := fmt.Fprintf(diffWriter, "Object with key %v not found", expectedUpdate.Key)
			if err != nil {
				panic(err)
			}
			eq = false
		}

		if !schematesting.CompareObjectKeys(expected.ObjectType().KeyFields, expectedUpdate.Key, actualUpdate.Key, diffWriter) {
			eq = false
		}

		return true
	})

	return eq
}
