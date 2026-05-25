package types_test

import (
	"fmt"
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
			"denom units at exact cap (Validate is bound-unaware)",
			buildMetadataWithDenomUnits(types.MaxDenomUnits),
			false,
		},
		{
			"over-cap denom units accepted by Validate (use ValidateBounds for the cap)",
			buildMetadataWithDenomUnits(types.MaxDenomUnits + 1),
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

// TestMetadataValidateBounds covers the write-time size guard introduced for
// #26012. Validate intentionally lets historically large stored metadata
// through (so genesis re-import keeps working); ValidateBounds is the entry
// point that callers on the write path should run to enforce the caps.
func TestMetadataValidateBounds(t *testing.T) {
	t.Run("within both caps", func(t *testing.T) {
		md := buildMetadataWithDenomUnits(types.MaxDenomUnits)
		md.DenomUnits[0].Aliases = buildAliases(types.MaxDenomUnitAliases)
		require.NoError(t, md.ValidateBounds())
	})

	t.Run("denom-unit count over cap", func(t *testing.T) {
		md := buildMetadataWithDenomUnits(types.MaxDenomUnits + 1)
		err := md.ValidateBounds()
		require.Error(t, err)
		require.Contains(t, err.Error(), "too many denom units")
	})

	t.Run("alias count over cap on inner unit", func(t *testing.T) {
		md := buildMetadataWithDenomUnits(2)
		md.DenomUnits[0].Aliases = buildAliases(types.MaxDenomUnitAliases + 1)
		err := md.ValidateBounds()
		require.Error(t, err)
		require.Contains(t, err.Error(), "too many aliases")
		require.Contains(t, err.Error(), md.DenomUnits[0].Denom)
	})
}

// buildMetadataWithDenomUnits returns a Metadata where DenomUnits has exactly
// `n` entries (base + n-1 ascending-exponent units) so the count-cap rule from
// #26012 is exercised without tripping the existing sort/duplicate checks.
func buildMetadataWithDenomUnits(n int) types.Metadata {
	units := make([]*types.DenomUnit, 0, n)
	units = append(units, &types.DenomUnit{Denom: "uatom", Exponent: 0})
	for i := 1; i < n; i++ {
		units = append(units, &types.DenomUnit{
			Denom:    fmt.Sprintf("atom%d", i),
			Exponent: uint32(i),
		})
	}
	// Last unit doubles as the display denom so the existing hasDisplay
	// invariant is satisfied for the cap-respecting case.
	display := "uatom"
	if n > 1 {
		display = units[n-1].Denom
	}
	return types.Metadata{
		Name:        "Cosmos Hub Atom",
		Symbol:      "ATOM",
		Description: "Synthetic metadata exercising the DenomUnits length cap.",
		DenomUnits:  units,
		Base:        "uatom",
		Display:     display,
	}
}

// buildAliases returns `n` distinct, non-blank alias strings.
func buildAliases(n int) []string {
	aliases := make([]string, n)
	for i := 0; i < n; i++ {
		aliases[i] = fmt.Sprintf("alias%d", i)
	}
	return aliases
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
