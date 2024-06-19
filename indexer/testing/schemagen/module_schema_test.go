package schemagen

import (
	"testing"

	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

func TestModuleSchema(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		schema := ModuleSchema.Draw(t, "schema")
		require.NoError(t, schema.Validate())
	})
}
