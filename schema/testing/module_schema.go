package schematesting

import (
	"fmt"

	"pgregory.net/rapid"

	"cosmossdk.io/schema"
)

var ModuleSchemaGen = rapid.Custom(func(t *rapid.T) schema.ModuleSchema {
	schema := schema.ModuleSchema{}
	numObjectTypes := rapid.IntRange(1, 10).Draw(t, "numObjectTypes")
	for i := 0; i < numObjectTypes; i++ {
		objectType := ObjectTypeGen.Draw(t, fmt.Sprintf("objectType[%d]", i))
		schema.ObjectTypes = append(schema.ObjectTypes, objectType)
	}
	return schema
}).Filter(func(schema schema.ModuleSchema) bool {
	// filter out enums with duplicate names
	enumTypeNames := map[string]bool{}
	for _, objectType := range schema.ObjectTypes {
		if !checkDuplicateEnumName(enumTypeNames, objectType.KeyFields) {
			return false
		}
		if !checkDuplicateEnumName(enumTypeNames, objectType.ValueFields) {
			return false
		}
	}
	return true
})
