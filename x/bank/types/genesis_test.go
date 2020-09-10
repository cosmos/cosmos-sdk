package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

func TestMarshalJSONMetaData(t *testing.T) {
	cdc := codec.NewLegacyAmino()

	testCases := []struct {
		name      string
		input     []types.Metadata
		strOutput string
	}{
		{"nil metadata", nil, `null`},
		{"empty metadata", []types.Metadata{}, `[]`},
		{"non-empty coins", []types.Metadata{{
			Description: "The native staking token of the Cosmos Hub.",
			DenomUnits: []*types.DenomUnit{
				{"uatom", uint32(0), []string{"microatom"}}, // The default exponent value 0 is omitted in the json
				{"matom", uint32(3), []string{"milliatom"}},
				{"atom", uint32(6), nil},
			},
			Base:    "uatom",
			Display: "atom",
		},
		},
			`[{"description":"The native staking token of the Cosmos Hub.","denom_units":[{"denom":"uatom","aliases":["microatom"]},{"denom":"matom","exponent":3,"aliases":["milliatom"]},{"denom":"atom","exponent":6}],"base":"uatom","display":"atom"}]`},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			bz, err := cdc.MarshalJSON(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.strOutput, string(bz))

			var newMetadata []types.Metadata
			require.NoError(t, cdc.UnmarshalJSON(bz, &newMetadata))

			if len(tc.input) == 0 {
				require.Nil(t, newMetadata)
			} else {
				require.Equal(t, tc.input, newMetadata)
			}
		})
	}
}
