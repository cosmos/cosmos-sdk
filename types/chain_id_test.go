package types

import (
	strings "strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseChainIDFromGenesis(t *testing.T) {
	chain_id, err := ParseChainIDFromGenesis(strings.NewReader(`{
		"chain-id":"test-chain-id",
		"state": {
		"accounts": [
		"abc": {},
		"efg": {},
		],
		},
	}`))
	require.NoError(t, err)
	require.Equal(t, "test-chain-id", chain_id)
}
