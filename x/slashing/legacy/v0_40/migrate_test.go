package v040_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v039slashing "github.com/cosmos/cosmos-sdk/x/slashing/legacy/v0_39"
	v040slashing "github.com/cosmos/cosmos-sdk/x/slashing/legacy/v0_40"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
)

func TestMigrate(t *testing.T) {
	v039Codec := codec.New()
	cryptocodec.RegisterCrypto(v039Codec)

	addr1, err := sdk.ConsAddressFromBech32("cosmosvalcons104cjmxkrg8y8lmrp25de02e4zf00zle4mzs685")
	require.NoError(t, err)
	addr2, err := sdk.ConsAddressFromBech32("cosmosvalcons10e4c5p6qk0sycy9u6u43t7csmlx9fyadr9yxph")
	require.NoError(t, err)

	gs := v039slashing.GenesisState{
		Params: types.DefaultParams(),
		SigningInfos: map[string]types.ValidatorSigningInfo{
			"cosmosvalcons10e4c5p6qk0sycy9u6u43t7csmlx9fyadr9yxph": {
				Address:             addr2,
				IndexOffset:         615501,
				MissedBlocksCounter: 1,
				Tombstoned:          false,
			},
			"cosmosvalcons104cjmxkrg8y8lmrp25de02e4zf00zle4mzs685": {
				Address:             addr1,
				IndexOffset:         2,
				MissedBlocksCounter: 2,
				Tombstoned:          false,
			},
		},
		MissedBlocks: map[string][]types.MissedBlock{
			"cosmosvalcons10e4c5p6qk0sycy9u6u43t7csmlx9fyadr9yxph": {
				{
					Index:  2,
					Missed: true,
				},
			},
			"cosmosvalcons104cjmxkrg8y8lmrp25de02e4zf00zle4mzs685": {
				{
					Index:  3,
					Missed: true,
				},
				{
					Index:  4,
					Missed: true,
				},
			},
		},
	}

	migrated := v040slashing.Migrate(gs)
	// Check that in `signing_infos` and `missed_blocks`, the address
	// cosmosvalcons104cjmxkrg8y8lmrp25de02e4zf00zle4mzs685
	// should always come before the address
	// cosmosvalcons10e4c5p6qk0sycy9u6u43t7csmlx9fyadr9yxph
	// (in alphabetic order, basically).
	expected := `{
  "params": {
    "signed_blocks_window": "100",
    "min_signed_per_window": "0.500000000000000000",
    "downtime_jail_duration": "600000000000",
    "slash_fraction_double_sign": "0.050000000000000000",
    "slash_fraction_downtime": "0.010000000000000000"
  },
  "signing_infos": [
    {
      "address": "cosmosvalcons104cjmxkrg8y8lmrp25de02e4zf00zle4mzs685",
      "validator_signing_info": {
        "address": "cosmosvalcons104cjmxkrg8y8lmrp25de02e4zf00zle4mzs685",
        "index_offset": "2",
        "jailed_until": "0001-01-01T00:00:00Z",
        "missed_blocks_counter": "2"
      }
    },
    {
      "address": "cosmosvalcons10e4c5p6qk0sycy9u6u43t7csmlx9fyadr9yxph",
      "validator_signing_info": {
        "address": "cosmosvalcons10e4c5p6qk0sycy9u6u43t7csmlx9fyadr9yxph",
        "index_offset": "615501",
        "jailed_until": "0001-01-01T00:00:00Z",
        "missed_blocks_counter": "1"
      }
    }
  ],
  "missed_blocks": [
    {
      "address": "cosmosvalcons104cjmxkrg8y8lmrp25de02e4zf00zle4mzs685",
      "missed_blocks": [
        {
          "index": "3",
          "missed": true
        },
        {
          "index": "4",
          "missed": true
        }
      ]
    },
    {
      "address": "cosmosvalcons10e4c5p6qk0sycy9u6u43t7csmlx9fyadr9yxph",
      "missed_blocks": [
        {
          "index": "2",
          "missed": true
        }
      ]
    }
  ]
}`

	bz, err := v039Codec.MarshalJSONIndent(migrated, "", "  ")
	require.NoError(t, err)
	require.Equal(t, expected, string(bz))
}
