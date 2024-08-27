package schematesting

import (
	"slices"

	"pgregory.net/rapid"

	"cosmossdk.io/schema"
)

// ModuleSchemaGen generates random ModuleSchema's based on the validity criteria of module schemas.
func ModuleSchemaGen() *rapid.Generator[schema.ModuleSchema] {
	enumTypesGen := distinctTypes(EnumType())
	return rapid.Custom(func(t *rapid.T) schema.ModuleSchema {
		enumTypes := enumTypesGen.Draw(t, "enumTypes")
		tempSchema, err := schema.NewModuleSchema(enumTypes...)
		if err != nil {
			t.Fatal(err)
		}
		objectTypes := distinctTypes(ObjectTypeGen(tempSchema)).Draw(t, "objectTypes")
		allTypes := append(enumTypes, objectTypes...)

		// remove duplicate type names
		slices.CompactFunc(allTypes, func(s schema.Type, s2 schema.Type) bool {
			return s.TypeName() == s2.TypeName()
		})

		modSchema, err := schema.NewModuleSchema(allTypes...)
		if err != nil {
			t.Fatal(err)
		}
		return modSchema
	})
}

func distinctTypes[T schema.Type](g *rapid.Generator[T]) *rapid.Generator[[]schema.Type] {
	return rapid.SliceOfNDistinct(rapid.Map(g, func(t T) schema.Type {
		return t
	}), 1, 10, func(t schema.Type) string {
		return t.TypeName()
	})
}
