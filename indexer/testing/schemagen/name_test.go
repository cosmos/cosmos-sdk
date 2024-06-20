package schemagen

import (
	"testing"

	indexerbase "cosmossdk.io/schema"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

func TestName(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		name := Name.Draw(t, "name")
		require.True(t, schema.ValidateName(name))
	})
}
