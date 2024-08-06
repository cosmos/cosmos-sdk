package schematesting

import (
	"pgregory.net/rapid"

	"cosmossdk.io/schema"
)

var enumValuesGen = rapid.SliceOfNDistinct(NameGen, 1, 10, func(x string) string { return x })

// EnumType generates random valid EnumTypes.
var EnumType = rapid.Custom(func(t *rapid.T) schema.EnumType {
	enum := schema.EnumType{
		Name:   NameGen.Draw(t, "name"),
		Values: enumValuesGen.Draw(t, "values"),
	}

	return enum
})
