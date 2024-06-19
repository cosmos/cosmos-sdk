package indexerrapid

import (
	"testing"

	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

func TestField(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		field := Field.Draw(t, "field")
		require.NoError(t, field.Validate())
	})
}

func TestFieldValue(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		field := Field.Draw(t, "field")
		require.NoError(t, field.Validate())
		fieldValue := FieldValue(field).Draw(t, "fieldValue")
		require.NoError(t, field.ValidateValue(fieldValue))
	})
}
