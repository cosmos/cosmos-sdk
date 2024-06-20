package schematesting

import (
	"testing"

	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

func TestObject(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		objectType := ObjectTypeGen.Draw(t, "object")
		require.NoError(t, objectType.Validate())
	})
}

func TestObjectUpdate(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		objectType := ObjectTypeGen.Draw(t, "object")
		require.NoError(t, objectType.Validate())
		update := ObjectInsertGen(objectType).Draw(t, "update")
		require.NoError(t, objectType.ValidateObjectUpdate(update))
	})
}

func TestExample(t *testing.T) {
	objectType := ObjectTypeGen.Example(1)
	update := ObjectInsertGen(objectType).Example(2)
	t.Logf("objectType: %+v", objectType)
	t.Logf("update: %+v", update)
}
