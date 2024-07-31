package schematesting

import (
	"testing"

	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

func TestAppSchema(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		schema := AppSchemaGen.Draw(t, "schema")
		for moduleName, moduleSchema := range schema {
			require.NotEmpty(t, moduleName)
			require.NoError(t, moduleSchema.Validate())
		}
	})
}
