package schematesting

import (
	"testing"

	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"

	"cosmossdk.io/schema"
)

func TestField(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		field := FieldGen(testEnumSchema).Draw(t, "field")
		require.NoError(t, field.Validate(testEnumSchema))
	})
}

func TestFieldValue(t *testing.T) {
	rapid.Check(t, checkFieldValue)
}

var checkFieldValue = func(t *rapid.T) {
	field := FieldGen(testEnumSchema).Draw(t, "field")
	require.NoError(t, field.Validate(testEnumSchema))
	fieldValue := FieldValueGen(field, testEnumSchema).Draw(t, "fieldValue")
	require.NoError(t, field.ValidateValue(fieldValue, testEnumSchema))
}

var testEnumSchema = schema.MustCompileModuleSchema(schema.EnumType{
	Name:   "test_enum",
	Values: []schema.EnumValueDefinition{{Name: "a", Value: 1}, {Name: "b", Value: 2}},
})
