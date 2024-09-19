package schematesting

import (
	"pgregory.net/rapid"

	"cosmossdk.io/schema"
)

// ModuleSchemaGen generates random ModuleSchema's based on the validity criteria of module schemas.
func ModuleSchemaGen() *rapid.Generator[schema.ModuleSchema] {
	enumTypesGen := distinctTypes(EnumType())
	return rapid.Custom(func(t *rapid.T) schema.ModuleSchema {
		enumTypes := enumTypesGen.Draw(t, "enumTypes")
		tempSchema, err := schema.CompileModuleSchema(enumTypes...)
		if err != nil {
			t.Fatal(err)
		}

		objectTypes := distinctTypes(StateObjectTypeGen(tempSchema)).Draw(t, "objectTypes")
		allTypes := append(enumTypes, objectTypes...)

		modSchema, err := schema.CompileModuleSchema(allTypes...)
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
