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
		capabilityOwners, ok := k.GetOwners(ctx, i)
		if !ok || len(capabilityOwners.Owners) == 0 {
			continue
		}

		genOwner := GenesisOwners{
			Index:  i,
			Owners: capabilityOwners,
		}
		owners = append(owners, genOwner)
	}

	return GenesisState{
		Index:  index,
		Owners: owners,
	}
}
