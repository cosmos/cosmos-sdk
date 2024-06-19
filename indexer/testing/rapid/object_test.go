package indexerrapid

import (
	"testing"

	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

func TestObject(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		objectType := ObjectType.Draw(t, "object")
		require.NoError(t, objectType.Validate())
	})
}

func TestObjectUpdate(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		objectType := ObjectType.Draw(t, "object")
		require.NoError(t, objectType.Validate())
		update := ObjectUpdate(objectType).Draw(t, "update")
		require.NoError(t, objectType.ValidateObjectUpdate(update))
	})
}
