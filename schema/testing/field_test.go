package schematesting

import (
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

func TestField(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		field := FieldGen.Draw(t, "field")
		require.NoError(t, field.Validate())
	})
}

func TestFieldValue(t *testing.T) {
	rapid.Check(t, checkFieldValue)
}

func FuzzFieldValue(f *testing.F) {
	strGen := rapid.String()
	f.Fuzz(rapid.MakeFuzz(func(t *rapid.T) {
		str := strGen.Draw(t, "str")
		if !utf8.ValidString(str) {
			t.Fatalf("invalid utf8 string: %q", str)
		}
	}))
}

var checkFieldValue = func(t *rapid.T) {
	field := FieldGen.Draw(t, "field")
	require.NoError(t, field.Validate())
	fieldValue := FieldValueGen(field).Draw(t, "fieldValue")
	require.NoError(t, field.ValidateValue(fieldValue))
}
