package indexerrapid

import (
	"fmt"

	"pgregory.net/rapid"

	indexerbase "cosmossdk.io/indexer/base"
)

var enumNumValues = rapid.IntRange(1, 25)

var EnumDefinition = rapid.Custom(func(t *rapid.T) indexerbase.EnumDefinition {
	num := enumNumValues.Draw(t, "numValues")
	values := make([]string, num)
	for i := 0; i < num; i++ {
		values[i] = nameGen.Draw(t, fmt.Sprintf("value[%d]", i))
	}
	enum := indexerbase.EnumDefinition{
		Name:   nameGen.Draw(t, "name"),
		Values: values,
	}

	return enum
})
