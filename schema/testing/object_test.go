package schematesting

import (
	"testing"

	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

func TestObject(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		objectType := ObjectTypeGen(testEnumSchema).Draw(t, "object")
		require.NoError(t, objectType.Validate(testEnumSchema))
	})
}

func TestObjectUpdate(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		objectType := ObjectTypeGen(testEnumSchema).Draw(t, "object")
		require.NoError(t, objectType.Validate(testEnumSchema))
		update := ObjectInsertGen(objectType, testEnumSchema).Draw(t, "update")
		require.NoError(t, objectType.ValidateObjectUpdate(update, testEnumSchema))
	})
}
