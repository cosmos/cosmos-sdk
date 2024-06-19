package indexerrapid

import (
	"pgregory.net/rapid"

	indexerbase "cosmossdk.io/indexer/base"
)

var enumValuesGen = rapid.SliceOfNDistinct(Name, 1, 10, func(x string) string { return x })

var EnumDefinition = rapid.Custom(func(t *rapid.T) indexerbase.EnumDefinition {
	enum := indexerbase.EnumDefinition{
		Name:   Name.Draw(t, "name"),
		Values: enumValuesGen.Draw(t, "values"),
	}

	return enum
})
