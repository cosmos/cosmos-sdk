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

// EnumType generates random valid EnumTypes.
func EnumType() *rapid.Generator[schema.EnumType] {
	return rapid.Custom(func(t *rapid.T) schema.EnumType {
		enum := schema.EnumType{
			Name:        NameGen.Draw(t, "name"),
			NumericKind: enumNumericKindGen.Draw(t, "numericKind"),
		}

		// we generate enum field values using FieldValueGen, which is a generator for random field values
		numericValueGen := FieldValueGen(schema.Field{Kind: enum.GetNumericKind()}, schema.EmptySchema{})
		numericValues := rapid.SliceOfNDistinct(rapid.Map(numericValueGen, func(a any) int32 { return a.(int32) }), 1, 10,
			func(a int32) int32 { return a }).Draw(t, "values")
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
