package schemagen

import (
	"github.com/tidwall/btree"
	"pgregory.net/rapid"

	indexerbase "cosmossdk.io/indexer/base"
)

var fieldsGen = rapid.SliceOfNDistinct(Field, 1, 12, func(f indexerbase.Field) string {
	return f.Name
})

var ObjectType = rapid.Custom(func(t *rapid.T) indexerbase.ObjectType {
	typ := indexerbase.ObjectType{
		Name: Name.Draw(t, "name"),
	}

	fields := fieldsGen.Draw(t, "fields")
	numKeyFields := rapid.IntRange(0, len(fields)).Draw(t, "numKeyFields")

	typ.KeyFields = fields[:numKeyFields]
	typ.ValueFields = fields[numKeyFields:]

	typ.RetainDeletions = boolGen.Draw(t, "retainDeletions")

	return typ
}).Filter(func(typ indexerbase.ObjectType) bool {
	// filter out duplicate enum names
	enumTypeNames := map[string]bool{}
	if !checkDuplicateEnumName(enumTypeNames, typ.KeyFields) {
		return false
	}
	if !checkDuplicateEnumName(enumTypeNames, typ.ValueFields) {
		return false
	}
	return true
})

func checkDuplicateEnumName(enumTypeNames map[string]bool, fields []indexerbase.Field) bool {
	for _, field := range fields {
		if field.Kind != indexerbase.EnumKind {
			continue
		}

		if _, ok := enumTypeNames[field.EnumDefinition.Name]; ok {
			return false
		}

		enumTypeNames[field.EnumDefinition.Name] = true
	}
	return true
}

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

func StatefulObjectUpdate(objectType indexerbase.ObjectType, state *btree.Map[string, indexerbase.ObjectUpdate]) *rapid.Generator[indexerbase.ObjectUpdate] {
	keyGen := KeyFieldsValue(objectType.KeyFields)
	valueGen := ValueFieldsValue(objectType.ValueFields)
	return rapid.Custom(func(t *rapid.T) indexerbase.ObjectUpdate {
		update := indexerbase.ObjectUpdate{
			TypeName: objectType.Name,
		}

		// TODO: when inserting a new object, all fields should be generated to avoid nil values
		// 50% of the time use existing key (when there are keys)
		n := state.Len()
		if n > 0 && boolGen.Draw(t, "existingKey") {
			i := rapid.IntRange(0, n-1).Draw(t, "index")
			update.Key = state.Values()[i].Key

			// delete 50% of the time
			if boolGen.Draw(t, "delete") {
				update.Delete = true
			} else {
				update.Value = valueGen.Draw(t, "value")
			}
		} else {
			update.Key = keyGen.Draw(t, "key")
			update.Value = valueGen.Draw(t, "value")
		}

		return update
	})
}
