package indexerrapid

import (
	"pgregory.net/rapid"

	indexerbase "cosmossdk.io/indexer/base"
)

var fieldsGen = rapid.SliceOfNDistinct(Field, 1, 12, func(f indexerbase.Field) string {
	return f.Name
})

var ObjectType = rapid.Custom(func(t *rapid.T) indexerbase.ObjectType {
	typ := indexerbase.ObjectType{
		Name: nameGen.Draw(t, "name"),
	}

	fields := fieldsGen.Draw(t, "fields")
	numKeyFields := rapid.IntRange(0, len(fields)).Draw(t, "numKeyFields")

	typ.KeyFields = fields[:numKeyFields]
	typ.ValueFields = fields[numKeyFields:]

	typ.RetainDeletions = boolGen.Draw(t, "retainDeletions")

	return typ
}).Filter(func(typ indexerbase.ObjectType) bool {
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
