package v047

import (
	"testing"

	"github.com/stretchr/testify/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

func TestMigrateGenState(t *testing.T) {
	tests := []struct {
		name     string
		oldState *types.GenesisState
		newState *types.GenesisState
	}{
		{
			name: "Balances supply metadata all unchanged",
			oldState: &types.GenesisState{
				Params: types.Params{},
				Balances: []types.Balance{{
					Address: "balance1",
					Coins:   sdk.Coins{sdk.NewCoin("balance1coin", sdk.NewInt(8))},
				}},
				Supply: sdk.Coins{sdk.NewCoin("supplycoin", sdk.NewInt(800))},
				DenomMetadata: []types.Metadata{{
					Description: "metadesk",
					DenomUnits:  nil,
					Base:        "meta",
					Display:     "meta",
					Name:        "foo",
					Symbol:      "META",
					URI:         "",
					URIHash:     "",
				}},
				SendEnabled: []types.SendEnabled{},
			},
			newState: &types.GenesisState{
				Params: types.Params{},
				Balances: []types.Balance{{
					Address: "balance1",
					Coins:   sdk.Coins{sdk.NewCoin("balance1coin", sdk.NewInt(8))},
				}},
				Supply: sdk.Coins{sdk.NewCoin("supplycoin", sdk.NewInt(800))},
				DenomMetadata: []types.Metadata{{
					Description: "metadesk",
					DenomUnits:  nil,
					Base:        "meta",
					Display:     "meta",
					Name:        "foo",
					Symbol:      "META",
					URI:         "",
					URIHash:     "",
				}},
				SendEnabled: []types.SendEnabled{},
			},
		},

		{
			name: "default send enabled true not changed",
			oldState: &types.GenesisState{
				Params: types.Params{DefaultSendEnabled: true},
			},
			newState: &types.GenesisState{
				Params: types.Params{DefaultSendEnabled: true},
			},
		},
		{
			name: "default send enabled false not changed",
			oldState: &types.GenesisState{
				Params: types.Params{DefaultSendEnabled: false, SendEnabled: []*types.SendEnabled{}},
			},
			newState: &types.GenesisState{
				Params: types.Params{DefaultSendEnabled: false},
			},
		},
		{
			name: "send enabled entries moved",
			oldState: &types.GenesisState{
				Params: types.Params{
					SendEnabled: []*types.SendEnabled{
						{"movecointrue", true},
						{"movecoinfalse", false},
					},
				},
			},
			newState: &types.GenesisState{
				Params: types.Params{},
				SendEnabled: []types.SendEnabled{
					{"movecointrue", true},
					{"movecoinfalse", false},
				},
			},
		},
		{
			name: "params entries added to existing",
			oldState: &types.GenesisState{
				Params: types.Params{
					SendEnabled: []*types.SendEnabled{
						{"movecointrue", true},
						{"movecoinfalse", false},
					},
				},
				SendEnabled: []types.SendEnabled{
					{"staycoin", true},
				},
			},
			newState: &types.GenesisState{
				Params: types.Params{},
				SendEnabled: []types.SendEnabled{
					{"staycoin", true},
					{"movecointrue", true},
					{"movecoinfalse", false},
				},
			},
		},
		{
			name: "conflicting params ignored",
			oldState: &types.GenesisState{
				Params: types.Params{
					SendEnabled: []*types.SendEnabled{
						{"staycoin", false},
					},
				},
				SendEnabled: []types.SendEnabled{
					{"staycoin", true},
				},
			},
			newState: &types.GenesisState{
				Params: types.Params{},
				SendEnabled: []types.SendEnabled{
					{"staycoin", true},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := MigrateGenState(tc.oldState)
			assert.Equal(t, tc.newState, actual)
		})
	}

	t.Run("ensure original not changed", func(t *testing.T) {
		origState := types.GenesisState{
			Params: types.Params{
				SendEnabled: []*types.SendEnabled{
					{"movecointrue", true},
					{"movecoinfalse", false},
				},
			},
		}
		_ = MigrateGenState(&origState)
		assert.Len(t, origState.Params.SendEnabled, 2)
	})
}
