package channel

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initializes the ibc channel submodule's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k Keeper, gs GenesisState) {
	for _, channel := range gs.Channels {
		k.SetChannel(ctx, channel)
	}
}

// ExportGenesis returns the ibc channel submodule's exported genesis.
func ExportGenesis(ctx sdk.Context, k Keeper) GenesisState {
	return GenesisState{
		Channels: k.GetAllChannels(ctx),
	}
}
