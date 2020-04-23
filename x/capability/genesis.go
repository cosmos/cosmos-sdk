package capability

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k Keeper, genState GenesisState) {
	k.SetIndex(ctx, genState.Index)

	// TODO: use cap index to set owners
	// for _, owner := range genState.Owners {

	// }
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k Keeper) GenesisState {
	return GenesisState{
		Index:  k.GetLatestIndex(ctx),
		Owners: nil, // TODO
	}
}
