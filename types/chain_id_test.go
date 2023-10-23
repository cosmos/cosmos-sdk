package types

import (
	_ "embed"
	"strings"
	"testing"

	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/stretchr/testify/require"
)

//go:embed testdata/streaming_chain_id.json
var BenchmarkGenesis string

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
				"chain_id": "test-chain-id"
			}`,
			"test-chain-id",
		},
		{
			"nested",
			`{
				"state": {
				  "accounts": {
					"a": {}
				  },
				  "chain_id": "test-chain-id"
				}
			}`,
			"",
		},
		{
			"not exist",
			`{
				"state": {
				  "accounts": {
					"a": {}
				  }
				},
				"chain-id": "test-chain-id"
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
		{
			"empty chain_id",
			`{"chain_id": ""}`,
			"",
		},
		{
			"chain_id too long",
			`{"chain_id": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"}`,
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

func BenchmarkParseChainID(b *testing.B) {
	b.ReportAllocs()
	b.Run("new", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			chainId, err := ParseChainIDFromGenesis(strings.NewReader(BenchmarkGenesis))
			require.NoError(b, err)
			require.Equal(b, "test_777-1", chainId)
		}
	})

	b.Run("old", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			doc, err := genutiltypes.AppGenesisFromReader(strings.NewReader(BenchmarkGenesis))
			require.NoError(b, err)
			require.Equal(b, "test_777-1", doc.ChainID)
		}
	})
}
