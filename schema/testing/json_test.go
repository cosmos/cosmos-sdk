package schematesting

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"

	"cosmossdk.io/schema"
)

func TestModuleSchemaJSON(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		modSchema := ModuleSchemaGen().Draw(t, "moduleSchema")
		bz, err := json.Marshal(modSchema)
		require.NoError(t, err)
		var modSchema2 schema.ModuleSchema
		err = json.Unmarshal(bz, &modSchema2)
		require.NoError(t, err)
		require.Equal(t, modSchema, modSchema2)
	})
}
