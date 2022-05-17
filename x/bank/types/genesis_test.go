package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestGenesisStateValidate(t *testing.T) {

	testCases := []struct {
		name         string
		genesisState GenesisState
		expErr       bool
	}{
		{
			"valid genesisState",
			GenesisState{
				Params: DefaultParams(),
				Balances: []Balance{
					{
						Address: "cosmos1yq8lgssgxlx9smjhes6ryjasmqmd3ts2559g0t",
						Coins:   sdk.Coins{sdk.NewInt64Coin("uatom", 1)},
					},
				},
				Supply: sdk.Coins{sdk.NewInt64Coin("uatom", 1)},
				DenomMetadata: []Metadata{
					{
						Name:        "Cosmos Hub Atom",
						Symbol:      "ATOM",
						Description: "The native staking token of the Cosmos Hub.",
						DenomUnits: []*DenomUnit{
							{"uatom", uint32(0), []string{"microatom"}},
							{"matom", uint32(3), []string{"milliatom"}},
							{"atom", uint32(6), nil},
						},
						Base:    "uatom",
						Display: "atom",
					},
				},
			},
			false,
		},
		{"empty genesisState", GenesisState{}, false},
		{
			"invalid params ",
			GenesisState{
				Params: Params{
					SendEnabled: []*SendEnabled{
						{"", true},
					},
				},
			},
			true,
		},
		{
			"dup balances",
			GenesisState{
				Balances: []Balance{
					{
						Address: "cosmos1yq8lgssgxlx9smjhes6ryjasmqmd3ts2559g0t",
						Coins:   sdk.Coins{sdk.NewInt64Coin("uatom", 1)},
					},
					{
						Address: "cosmos1yq8lgssgxlx9smjhes6ryjasmqmd3ts2559g0t",
						Coins:   sdk.Coins{sdk.NewInt64Coin("uatom", 1)},
					},
				},
			},
			true,
		},
		{
			"0  balance",
			GenesisState{
				Balances: []Balance{
					{
						Address: "cosmos1yq8lgssgxlx9smjhes6ryjasmqmd3ts2559g0t",
					},
				},
			},
			false,
		},
		{
			"dup Metadata",
			GenesisState{
				DenomMetadata: []Metadata{
					{
						Name:        "Cosmos Hub Atom",
						Symbol:      "ATOM",
						Description: "The native staking token of the Cosmos Hub.",
						DenomUnits: []*DenomUnit{
							{"uatom", uint32(0), []string{"microatom"}},
							{"matom", uint32(3), []string{"milliatom"}},
							{"atom", uint32(6), nil},
						},
						Base:    "uatom",
						Display: "atom",
					},
					{
						Name:        "Cosmos Hub Atom",
						Symbol:      "ATOM",
						Description: "The native staking token of the Cosmos Hub.",
						DenomUnits: []*DenomUnit{
							{"uatom", uint32(0), []string{"microatom"}},
							{"matom", uint32(3), []string{"milliatom"}},
							{"atom", uint32(6), nil},
						},
						Base:    "uatom",
						Display: "atom",
					},
				},
			},
			true,
		},
		{
			"invalid Metadata",
			GenesisState{
				DenomMetadata: []Metadata{
					{
						Name:        "Cosmos Hub Atom",
						Symbol:      "ATOM",
						Description: "The native staking token of the Cosmos Hub.",
						DenomUnits: []*DenomUnit{
							{"uatom", uint32(0), []string{"microatom"}},
							{"matom", uint32(3), []string{"milliatom"}},
							{"atom", uint32(6), nil},
						},
						Base:    "",
						Display: "",
					},
				},
			},
			true,
		},
		{
			"invalid supply",
			GenesisState{
				Supply: sdk.Coins{sdk.Coin{Denom: "", Amount: sdk.OneInt()}},
			},
			true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			err := tc.genesisState.Validate()

			if tc.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMigrateSendEnabled(t *testing.T) {
	tests := []struct {
		name     string
		oldState *GenesisState
		newState *GenesisState
	}{
		{
			name: "Balances supply metadata all unchanged",
			oldState: &GenesisState{
				Params: Params{},
				Balances: []Balance{{
					Address: "balance1",
					Coins:   sdk.Coins{sdk.NewCoin("balance1coin", sdk.NewInt(8))},
				}},
				Supply: sdk.Coins{sdk.NewCoin("supplycoin", sdk.NewInt(800))},
				DenomMetadata: []Metadata{{
					Description: "metadesk",
					DenomUnits:  nil,
					Base:        "meta",
					Display:     "meta",
					Name:        "foo",
					Symbol:      "META",
					URI:         "",
					URIHash:     "",
				}},
			},
			newState: &GenesisState{
				Params: Params{},
				Balances: []Balance{{
					Address: "balance1",
					Coins:   sdk.Coins{sdk.NewCoin("balance1coin", sdk.NewInt(8))},
				}},
				Supply: sdk.Coins{sdk.NewCoin("supplycoin", sdk.NewInt(800))},
				DenomMetadata: []Metadata{{
					Description: "metadesk",
					DenomUnits:  nil,
					Base:        "meta",
					Display:     "meta",
					Name:        "foo",
					Symbol:      "META",
					URI:         "",
					URIHash:     "",
				}},
			},
		},

		{
			name: "default send enabled true not changed",
			oldState: &GenesisState{
				Params: Params{DefaultSendEnabled: true},
			},
			newState: &GenesisState{
				Params: Params{DefaultSendEnabled: true},
			},
		},
		{
			name: "default send enabled false not changed",
			oldState: &GenesisState{
				Params: Params{DefaultSendEnabled: false, SendEnabled: []*SendEnabled{}},
			},
			newState: &GenesisState{
				Params: Params{DefaultSendEnabled: false},
			},
		},
		{
			name: "send enabled entries moved",
			oldState: &GenesisState{
				Params: Params{
					SendEnabled: []*SendEnabled{
						{"movecointrue", true},
						{"movecoinfalse", false},
					},
				},
			},
			newState: &GenesisState{
				Params: Params{},
				SendEnabled: []SendEnabled{
					{"movecointrue", true},
					{"movecoinfalse", false},
				},
			},
		},
		{
			name: "params entries added to existing",
			oldState: &GenesisState{
				Params: Params{
					SendEnabled: []*SendEnabled{
						{"movecointrue", true},
						{"movecoinfalse", false},
					},
				},
				SendEnabled: []SendEnabled{
					{"staycoin", true},
				},
			},
			newState: &GenesisState{
				Params: Params{},
				SendEnabled: []SendEnabled{
					{"staycoin", true},
					{"movecointrue", true},
					{"movecoinfalse", false},
				},
			},
		},
		{
			name: "conflicting params ignored",
			oldState: &GenesisState{
				Params: Params{
					SendEnabled: []*SendEnabled{
						{"staycoin", false},
					},
				},
				SendEnabled: []SendEnabled{
					{"staycoin", true},
				},
			},
			newState: &GenesisState{
				Params: Params{},
				SendEnabled: []SendEnabled{
					{"staycoin", true},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.oldState.MigrateSendEnabled()
			assert.Equal(t, tc.newState, tc.oldState)
		})
	}
}
