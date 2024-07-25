package schematesting

import (
	"fmt"

	"pgregory.net/rapid"

	"cosmossdk.io/schema"
)

// ModuleSchemaGen generates random ModuleSchema's based on the validity criteria of module schemas.
var ModuleSchemaGen = rapid.Custom(func(t *rapid.T) schema.ModuleSchema {
	objectTypes := objectTypesGen.Draw(t, "objectTypes")
	modSchema, err := schema.NewModuleSchema(objectTypes)
	if err != nil {
		t.Fatal(err)
	}
	return modSchema
})

var objectTypesGen = rapid.Custom(func(t *rapid.T) []schema.ObjectType {
	var objectTypes []schema.ObjectType
	numObjectTypes := rapid.IntRange(1, 10).Draw(t, "numObjectTypes")
	for i := 0; i < numObjectTypes; i++ {
		objectType := ObjectTypeGen.Draw(t, fmt.Sprintf("objectType[%d]", i))
		objectTypes = append(objectTypes, objectType)
	}
	return objectTypes
}).Filter(func(objectTypes []schema.ObjectType) bool {
	typeNames := map[string]bool{}
	for _, objectType := range objectTypes {
		if hasDuplicateNames(typeNames, objectType.KeyFields) || hasDuplicateNames(typeNames, objectType.ValueFields) {
			return false
		}
		if typeNames[objectType.Name] {
			return false
		}
		typeNames[objectType.Name] = true
	}
	return true
})

// MustNewModuleSchema calls NewModuleSchema and panics if there's an error. This should generally be used
// only in tests or initialization code.
func MustNewModuleSchema(objectTypes []schema.ObjectType) schema.ModuleSchema {
	schema, err := schema.NewModuleSchema(objectTypes)
	if err != nil {
		panic(err)
	}
	return schema
}
