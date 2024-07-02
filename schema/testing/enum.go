package schematesting

import (
	"pgregory.net/rapid"

	"cosmossdk.io/schema"
)

var enumValuesGen = rapid.SliceOfNDistinct(NameGen, 1, 10, func(x string) string { return x })

var EnumDefinitionGen = rapid.Custom(func(t *rapid.T) schema.EnumDefinition {
	enum := schema.EnumDefinition{
		Name:   NameGen.Draw(t, "name"),
		Values: enumValuesGen.Draw(t, "values"),
	}

	return enum
})
