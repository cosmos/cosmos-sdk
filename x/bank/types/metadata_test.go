package types_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

func TestMetadataValidate(t *testing.T) {
	testCases := []struct {
		name     string
		metadata types.Metadata
		expErr   bool
	}{
		{
			"non-empty coins",
			types.Metadata{
				Name:        "Cosmos Hub Atom",
				Symbol:      "ATOM",
				Description: "The native staking token of the Cosmos Hub.",
				DenomUnits: []*types.DenomUnit{
					{"uatom", uint32(0), []string{"microatom"}},
					{"matom", uint32(3), []string{"milliatom"}},
					{"atom", uint32(6), nil},
				},
				Base:    "uatom",
				Display: "atom",
			},
			false,
		},
		{
			"base coin is display coin",
			types.Metadata{
				Name:        "Cosmos Hub Atom",
				Symbol:      "ATOM",
				Description: "The native staking token of the Cosmos Hub.",
				DenomUnits: []*types.DenomUnit{
					{"atom", uint32(0), []string{"ATOM"}},
				},
				Base:    "atom",
				Display: "atom",
			},
			false,
		},
		{"empty metadata", types.Metadata{}, true},
		{
			"blank name",
			types.Metadata{
				Name: "",
			},
			true,
		},
		{
			"blank symbol",
			types.Metadata{
				Name:   "Cosmos Hub Atom",
				Symbol: "",
			},
			true,
		},
		{
			"invalid base denom",
			types.Metadata{
				Name:   "Cosmos Hub Atom",
				Symbol: "ATOM",
				Base:   "",
			},
			true,
		},
		{
			"invalid display denom",
			types.Metadata{
				Name:    "Cosmos Hub Atom",
				Symbol:  "ATOM",
				Base:    "uatom",
				Display: "",
			},
			true,
		},
		{
			"duplicate denom unit",
			types.Metadata{
				Name:        "Cosmos Hub Atom",
				Symbol:      "ATOM",
				Description: "The native staking token of the Cosmos Hub.",
				DenomUnits: []*types.DenomUnit{
					{"uatom", uint32(0), []string{"microatom"}},
					{"uatom", uint32(1), []string{"microatom"}},
				},
				Base:    "uatom",
				Display: "atom",
			},
			true,
		},
		{
			"invalid denom unit",
			types.Metadata{
				Name:        "Cosmos Hub Atom",
				Symbol:      "ATOM",
				Description: "The native staking token of the Cosmos Hub.",
				DenomUnits: []*types.DenomUnit{
					{"", uint32(0), []string{"microatom"}},
				},
				Base:    "uatom",
				Display: "atom",
			},
			true,
		},
		{
			"invalid denom unit alias",
			types.Metadata{
				Name:        "Cosmos Hub Atom",
				Symbol:      "ATOM",
				Description: "The native staking token of the Cosmos Hub.",
				DenomUnits: []*types.DenomUnit{
					{"uatom", uint32(0), []string{""}},
				},
				Base:    "uatom",
				Display: "atom",
			},
			true,
		},
		{
			"duplicate denom unit alias",
			types.Metadata{
				Name:        "Cosmos Hub Atom",
				Symbol:      "ATOM",
				Description: "The native staking token of the Cosmos Hub.",
				DenomUnits: []*types.DenomUnit{
					{"uatom", uint32(0), []string{"microatom", "microatom"}},
				},
				Base:    "uatom",
				Display: "atom",
			},
			true,
		},
		{
			"no base denom unit",
			types.Metadata{
				Name:        "Cosmos Hub Atom",
				Symbol:      "ATOM",
				Description: "The native staking token of the Cosmos Hub.",
				DenomUnits: []*types.DenomUnit{
					{"matom", uint32(3), []string{"milliatom"}},
					{"atom", uint32(6), nil},
				},
				Base:    "uatom",
				Display: "atom",
			},
			true,
		},
		{
			"base denom exponent not zero",
			types.Metadata{
				Name:        "Cosmos Hub Atom",
				Symbol:      "ATOM",
				Description: "The native staking token of the Cosmos Hub.",
				DenomUnits: []*types.DenomUnit{
					{"uatom", uint32(1), []string{"microatom"}},
					{"matom", uint32(3), []string{"milliatom"}},
					{"atom", uint32(6), nil},
				},
				Base:    "uatom",
				Display: "atom",
			},
			true,
		},
		{
			"invalid denom unit",
			types.Metadata{
				Name:        "Cosmos Hub Atom",
				Symbol:      "ATOM",
				Description: "The native staking token of the Cosmos Hub.",
				DenomUnits: []*types.DenomUnit{
					{"uatom", uint32(0), []string{"microatom"}},
					{"", uint32(3), []string{"milliatom"}},
				},
				Base:    "uatom",
				Display: "uatom",
			},
			true,
		},
		{
			"no display denom unit",
			types.Metadata{
				Name:        "Cosmos Hub Atom",
				Symbol:      "ATOM",
				Description: "The native staking token of the Cosmos Hub.",
				DenomUnits: []*types.DenomUnit{
					{"uatom", uint32(0), []string{"microatom"}},
				},
				Base:    "uatom",
				Display: "atom",
			},
			true,
		},
		{
			"denom units not sorted",
			types.Metadata{
				Name:        "Cosmos Hub Atom",
				Symbol:      "ATOM",
				Description: "The native staking token of the Cosmos Hub.",
				DenomUnits: []*types.DenomUnit{
					{"uatom", uint32(0), []string{"microatom"}},
					{"atom", uint32(6), nil},
					{"matom", uint32(3), []string{"milliatom"}},
				},
				Base:    "uatom",
				Display: "atom",
			},
			true,
		},
		{
			"too many denom units",
			types.Metadata{
				Name:        "Test Token",
				Symbol:      "TEST",
				Description: "Token with too many denom units.",
				DenomUnits: func() []*types.DenomUnit {
					units := make([]*types.DenomUnit, types.MaxDenomUnits+1)
					for i := 0; i <= types.MaxDenomUnits; i++ {
						units[i] = &types.DenomUnit{Denom: fmt.Sprintf("unit%d", i), Exponent: uint32(i)}
					}
					return units
				}(),
				Base:    "unit0",
				Display: fmt.Sprintf("unit%d", types.MaxDenomUnits),
			},
			true,
		},
		{
			"max denom units",
			types.Metadata{
				Name:        "Test Token",
				Symbol:      "TEST",
				Description: "Token with max allowed denom units.",
				DenomUnits: func() []*types.DenomUnit {
					units := make([]*types.DenomUnit, types.MaxDenomUnits)
					for i := 0; i < types.MaxDenomUnits; i++ {
						units[i] = &types.DenomUnit{Denom: fmt.Sprintf("unit%d", i), Exponent: uint32(i)}
					}
					return units
				}(),
				Base:    "unit0",
				Display: fmt.Sprintf("unit%d", types.MaxDenomUnits-1),
			},
			false,
		},
		{
			"too many aliases",
			types.Metadata{
				Name:        "Cosmos Hub Atom",
				Symbol:      "ATOM",
				Description: "The native staking token of the Cosmos Hub.",
				DenomUnits: []*types.DenomUnit{
					{
						Denom:    "uatom",
						Exponent: 0,
						Aliases: func() []string {
							aliases := make([]string, types.MaxAliases+1)
							for i := 0; i <= types.MaxAliases; i++ {
								aliases[i] = fmt.Sprintf("alias%d", i)
							}
							return aliases
						}(),
					},
				},
				Base:    "uatom",
				Display: "uatom",
			},
			true,
		},
		{
			"max aliases",
			types.Metadata{
				Name:        "Cosmos Hub Atom",
				Symbol:      "ATOM",
				Description: "The native staking token of the Cosmos Hub.",
				DenomUnits: []*types.DenomUnit{
					{
						Denom:    "uatom",
						Exponent: 0,
						Aliases: func() []string {
							aliases := make([]string, types.MaxAliases)
							for i := 0; i < types.MaxAliases; i++ {
								aliases[i] = fmt.Sprintf("alias%d", i)
							}
							return aliases
						}(),
					},
				},
				Base:    "uatom",
				Display: "uatom",
			},
			false,
		},
		{
			"alias too long",
			types.Metadata{
				Name:        "Cosmos Hub Atom",
				Symbol:      "ATOM",
				Description: "The native staking token of the Cosmos Hub.",
				DenomUnits: []*types.DenomUnit{
					{
						Denom:    "uatom",
						Exponent: 0,
						Aliases:  []string{strings.Repeat("a", types.MaxAliasLength+1)},
					},
				},
				Base:    "uatom",
				Display: "uatom",
			},
			true,
		},
		{
			"max alias length",
			types.Metadata{
				Name:        "Cosmos Hub Atom",
				Symbol:      "ATOM",
				Description: "The native staking token of the Cosmos Hub.",
				DenomUnits: []*types.DenomUnit{
					{
						Denom:    "uatom",
						Exponent: 0,
						Aliases:  []string{strings.Repeat("a", types.MaxAliasLength)},
					},
				},
				Base:    "uatom",
				Display: "uatom",
			},
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.metadata.Validate()

			if tc.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMarshalJSONMetaData(t *testing.T) {
	cdc := codec.NewLegacyAmino()

	testCases := []struct {
		name      string
		input     []types.Metadata
		strOutput string
	}{
		{"nil metadata", nil, `null`},
		{"empty metadata", []types.Metadata{}, `[]`},
		{
			"non-empty coins",
			[]types.Metadata{
				{
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
			`[{"description":"The native staking token of the Cosmos Hub.","denom_units":[{"denom":"uatom","aliases":["microatom"]},{"denom":"matom","exponent":3,"aliases":["milliatom"]},{"denom":"atom","exponent":6}],"base":"uatom","display":"atom"}]`,
		},
	}

	for _, tc := range testCases {
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
