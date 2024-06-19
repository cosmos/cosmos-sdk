package indexerrapid

import (
	"fmt"

	"pgregory.net/rapid"

	indexerbase "cosmossdk.io/indexer/base"
)

var numKeyFields = rapid.IntRange(0, 5)
var numValueFields = rapid.IntRange(0, 10)

var ObjectType = rapid.Custom(func(t *rapid.T) indexerbase.ObjectType {
	typ := indexerbase.ObjectType{
		Name: nameGen.Draw(t, "name"),
	}

	numKey := numKeyFields.Draw(t, "numKeyFields")
	typ.KeyFields = make([]indexerbase.Field, numKey)
	for i := 0; i < numKey; i++ {
		typ.KeyFields[i] = Field.Draw(t, fmt.Sprintf("keyField[%d]", i))
	}

	numValue := numValueFields.Draw(t, "numValueFields")
	typ.ValueFields = make([]indexerbase.Field, numValue)
	for i := 0; i < numValue; i++ {
		typ.ValueFields[i] = Field.Draw(t, fmt.Sprintf("valueField[%d]", i))
	}

	typ.RetainDeletions = boolGen.Draw(t, "retainDeletions")

	return typ
}).Filter(func(typ indexerbase.ObjectType) bool {
	// TODO: filter out empty key & value fields
	// TODO: filter out duplicate field names
	// TODO: filter out duplicate enum names
	return true
})

func ObjectUpdate(objectType indexerbase.ObjectType) *rapid.Generator[indexerbase.ObjectUpdate] {
	keyGen := KeyFieldsValue(objectType.KeyFields)
	valueGen := ValueFieldsValue(objectType.ValueFields)
	return rapid.Custom(func(t *rapid.T) indexerbase.ObjectUpdate {
		update := indexerbase.ObjectUpdate{
			TypeName: objectType.Name,
		}

		update.Key = keyGen.Draw(t, "key")

		// delete 50% of the time
		if boolGen.Draw(t, "delete") {
			update.Delete = true
		} else {
			update.Value = valueGen.Draw(t, "value")
		}

		return update
	})
}
