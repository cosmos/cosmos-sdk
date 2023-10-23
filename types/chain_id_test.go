package types

import (
	strings "strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseChainIDFromGenesis(t *testing.T) {
	testCases := []struct {
		name       string
		json       string
		expChainID string
	}{
		{
			"success",
			`{
				"state": {
				  "accounts": {
					"a": {}
				  }
				},
				"chain-id": "test-chain-id"
			}`,
			"test-chain-id",
		},
		{
			"not exist",
			`{
				"state": {
				  "accounts": {
					"a": {}
				  }
				},
				"chain_id": "test-chain-id"
			}`,
			"",
		},
		{
			"invalid type",
			`{
				"chain-id":1,
			}`,
			"",
		},
		{
			"invalid json",
			`[ " ': }`,
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			chain_id, err := ParseChainIDFromGenesis(strings.NewReader(tc.json))
			if tc.expChainID == "" {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expChainID, chain_id)
			}
		})
	}
}
