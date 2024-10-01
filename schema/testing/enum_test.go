package schematesting

import (
	"testing"

	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"

	"cosmossdk.io/schema"
)

func TestEnumType(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		enumType := EnumType().Draw(t, "enum")
		require.NoError(t, enumType.Validate(schema.EmptyTypeSet()))
	})
}
