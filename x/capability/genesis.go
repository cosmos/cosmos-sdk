package capability

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k Keeper, genState GenesisState) {
	k.SetIndex(ctx, genState.Index)

	// set owners for each index and initialize capability
	for _, genOwner := range genState.Owners {
		k.SetOwners(ctx, genOwner.Index, genOwner.Owners)
		k.InitializeCapability(ctx, genOwner.Index, genOwner.Owners)
	}
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k Keeper) GenesisState {
	index := k.GetLatestIndex(ctx)
	owners := []GenesisOwners{}
	for i := uint64(1); i < index; i++ {
		iOwners, ok := k.GetOwners(ctx, i)

		// only export non-empty owners
		if ok && len(iOwners.Owners) != 0 {
			genOwner := GenesisOwners{
				Index:  i,
				Owners: *iOwners,
			}
			owners = append(owners, genOwner)
		}
	}
	return GenesisState{
		Index:  index,
		Owners: owners,
	}
}
