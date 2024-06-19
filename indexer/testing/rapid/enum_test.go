package indexerrapid

import (
	"testing"

	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

func TestEnumDefinition(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		enumDefinition := EnumDefinition.Draw(t, "enum")
		require.NoError(t, enumDefinition.Validate())
	})
}
