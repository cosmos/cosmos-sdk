package schemagen

import (
	"testing"

	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"

	"cosmossdk.io/schema"
)

func TestName(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		name := Name.Draw(t, "name")
		require.True(t, schema.ValidateName(name))
	})
}
