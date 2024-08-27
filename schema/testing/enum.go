package schematesting

import (
	"pgregory.net/rapid"

	"cosmossdk.io/schema"
)

var enumNumericKindGen = rapid.SampledFrom([]schema.Kind{
	schema.InvalidKind,
	schema.Int8Kind,
	schema.Int16Kind,
	schema.Int32Kind,
	schema.Uint8Kind,
	schema.Uint16Kind,
})

var enumNumericValueGens = map[schema.Kind]*rapid.Generator[int32]{
	schema.Int8Kind:   rapid.Map(rapid.Int8(), func(a int8) int32 { return int32(a) }),
	schema.Int16Kind:  rapid.Map(rapid.Int16(), func(a int16) int32 { return int32(a) }),
	schema.Int32Kind:  rapid.Map(rapid.Int32(), func(a int32) int32 { return a }),
	schema.Uint8Kind:  rapid.Map(rapid.Uint8(), func(a uint8) int32 { return int32(a) }),
	schema.Uint16Kind: rapid.Map(rapid.Uint16(), func(a uint16) int32 { return int32(a) }),
}

// EnumType generates random valid EnumTypes.
func EnumType() *rapid.Generator[schema.EnumType] {
	return rapid.Custom(func(t *rapid.T) schema.EnumType {
		enum := schema.EnumType{
			Name:        NameGen.Draw(t, "name"),
			NumericKind: enumNumericKindGen.Draw(t, "numericKind"),
		}

		numericValueGen := enumNumericValueGens[enum.GetNumericKind()]
		numericValues := rapid.SliceOfNDistinct(numericValueGen, 1, 10, func(e int32) int32 {
			return e
		}).Draw(t, "values")
		n := len(numericValues)
		names := rapid.SliceOfNDistinct(NameGen, n, n, func(a string) string { return a }).Draw(t, "names")
		values := make([]schema.EnumValueDefinition, n)
		for i, v := range numericValues {
			values[i] = schema.EnumValueDefinition{Name: names[i], Value: v}
		}

		enum.Values = values

		return enum
	})
}
