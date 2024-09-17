package schematesting

import (
	"github.com/tidwall/btree"
	"pgregory.net/rapid"

	"cosmossdk.io/schema"
)

// StateObjectTypeGen generates random StateObjectType's based on the validity criteria of object types.
func StateObjectTypeGen(typeSet schema.TypeSet) *rapid.Generator[schema.StateObjectType] {
	keyFieldsGen := rapid.SliceOfNDistinct(KeyFieldGen(typeSet), 1, 6, func(f schema.Field) string {
		return f.Name
	})

	valueFieldsGen := rapid.SliceOfNDistinct(FieldGen(typeSet), 1, 12, func(f schema.Field) string {
		return f.Name
	})

	return rapid.Custom(func(t *rapid.T) schema.StateObjectType {
		typ := schema.StateObjectType{
			Name: NameGen.Filter(func(s string) bool {
				// filter out names that already exist in the schema
				_, found := typeSet.LookupType(s)
				return !found
			}).Draw(t, "name"),
		}

		typ.KeyFields = keyFieldsGen.Draw(t, "keyFields")
		typ.ValueFields = valueFieldsGen.Draw(t, "valueFields")
		typ.RetainDeletions = boolGen.Draw(t, "retainDeletions")

		return typ
	}).Filter(func(typ schema.StateObjectType) bool {
		// filter out duplicate field names
		fieldNames := map[string]bool{}
		if hasDuplicateFieldNames(fieldNames, typ.KeyFields) {
			return false
		}
		if hasDuplicateFieldNames(fieldNames, typ.ValueFields) {
			return false
		}

		return true
	})
}

func hasDuplicateFieldNames(typeNames map[string]bool, fields []schema.Field) bool {
	for _, field := range fields {
		if _, ok := typeNames[field.Name]; ok {
			return true
		}
		typeNames[field.Name] = true
	}
	return false
}

// StateObjectInsertGen generates object updates that are valid for insertion.
func StateObjectInsertGen(objectType schema.StateObjectType, typeSet schema.TypeSet) *rapid.Generator[schema.StateObjectUpdate] {
	return StateObjectUpdateGen(objectType, nil, typeSet)
}

// StateObjectUpdateGen generates object updates that are valid for updates using the provided state map as a source
// of valid existing keys.
func StateObjectUpdateGen(objectType schema.StateObjectType, state *btree.Map[string, schema.StateObjectUpdate], sch schema.TypeSet) *rapid.Generator[schema.StateObjectUpdate] {
	keyGen := ObjectKeyGen(objectType.KeyFields, sch).Filter(func(key interface{}) bool {
		// filter out keys that exist in the state
		if state != nil {
			_, exists := state.Get(ObjectKeyString(objectType, key))
			return !exists
		}
		return true
	})
	insertValueGen := ObjectValueGen(objectType.ValueFields, false, sch)
	updateValueGen := ObjectValueGen(objectType.ValueFields, true, sch)
	return rapid.Custom(func(t *rapid.T) schema.StateObjectUpdate {
		update := schema.StateObjectUpdate{
			TypeName: objectType.Name,
		}

		// 50% of the time use existing key (when there are keys)
		n := 0
		if state != nil {
			n = state.Len()
		}
		if n > 0 && boolGen.Draw(t, "existingKey") {
			i := rapid.IntRange(0, n-1).Draw(t, "index")
			update.Key = state.Values()[i].Key

			// delete 50% of the time
			if boolGen.Draw(t, "delete") {
				update.Delete = true
			} else {
				update.Value = updateValueGen.Draw(t, "value")
			}
		} else {
			update.Key = keyGen.Draw(t, "key")
			update.Value = insertValueGen.Draw(t, "value")
		}

		return update
	})
}
