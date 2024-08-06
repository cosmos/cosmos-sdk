package schematesting

import (
	"github.com/tidwall/btree"
	"pgregory.net/rapid"

	"cosmossdk.io/schema"
)

var fieldsGen = rapid.SliceOfNDistinct(FieldGen, 1, 12, func(f schema.Field) string {
	return f.Name
})

// ObjectTypeGen generates random ObjectType's based on the validity criteria of object types.
var ObjectTypeGen = rapid.Custom(func(t *rapid.T) schema.ObjectType {
	typ := schema.ObjectType{
		Name: NameGen.Draw(t, "name"),
	}

	fields := fieldsGen.Draw(t, "fields")
	numKeyFields := rapid.IntRange(0, len(fields)).Draw(t, "numKeyFields")

	typ.KeyFields = fields[:numKeyFields]

	for i := range typ.KeyFields {
		// key fields can't be nullable
		typ.KeyFields[i].Nullable = false
	}

	typ.ValueFields = fields[numKeyFields:]

	typ.RetainDeletions = boolGen.Draw(t, "retainDeletions")

	return typ
}).Filter(func(typ schema.ObjectType) bool {
	// filter out duplicate enum names
	typeNames := map[string]bool{typ.Name: true}
	if hasDuplicateNames(typeNames, typ.KeyFields) {
		return false
	}
	if hasDuplicateNames(typeNames, typ.ValueFields) {
		return false
	}
	return true
})

// hasDuplicateNames checks if there is type name in the fields
func hasDuplicateNames(typeNames map[string]bool, fields []schema.Field) bool {
	for _, field := range fields {
		if field.Kind != schema.EnumKind {
			continue
		}

		if _, ok := typeNames[field.EnumType.Name]; ok {
			return true
		}

		typeNames[field.EnumType.Name] = true
	}
	return false
}

// ObjectInsertGen generates object updates that are valid for insertion.
func ObjectInsertGen(objectType schema.ObjectType) *rapid.Generator[schema.ObjectUpdate] {
	return ObjectUpdateGen(objectType, nil)
}

// ObjectUpdateGen generates object updates that are valid for updates using the provided state map as a source
// of valid existing keys.
func ObjectUpdateGen(objectType schema.ObjectType, state *btree.Map[string, schema.ObjectUpdate]) *rapid.Generator[schema.ObjectUpdate] {
	keyGen := ObjectKeyGen(objectType.KeyFields)

	if len(objectType.ValueFields) == 0 {
		// special case where there are no value fields,
		// so we just insert or delete, no updates
		return rapid.Custom(func(t *rapid.T) schema.ObjectUpdate {
			update := schema.ObjectUpdate{
				TypeName: objectType.Name,
			}

			// 50% of the time delete existing key (when there are keys)
			n := 0
			if state != nil {
				n = state.Len()
			}
			if n > 0 && boolGen.Draw(t, "delete") {
				i := rapid.IntRange(0, n-1).Draw(t, "index")
				update.Key = state.Values()[i].Key
				update.Delete = true
			} else {
				update.Key = keyGen.Draw(t, "key")
			}

			return update
		})
	} else {
		insertValueGen := ObjectValueGen(objectType.ValueFields, false)
		updateValueGen := ObjectValueGen(objectType.ValueFields, true)
		return rapid.Custom(func(t *rapid.T) schema.ObjectUpdate {
			update := schema.ObjectUpdate{
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
}
