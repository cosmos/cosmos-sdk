package capability

import (
	sdk "github.com/KiraCore/cosmos-sdk/types"
	"github.com/KiraCore/cosmos-sdk/x/capability/keeper"
	"github.com/KiraCore/cosmos-sdk/x/capability/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	k.SetIndex(ctx, genState.Index)

	// set owners for each index and initialize capability
	for _, genOwner := range genState.Owners {
		k.SetOwners(ctx, genOwner.Index, genOwner.Owners)
		k.InitializeCapability(ctx, genOwner.Index, genOwner.Owners)
	}
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) types.GenesisState {
	index := k.GetLatestIndex(ctx)
	owners := []types.GenesisOwners{}

	for i := uint64(1); i < index; i++ {
		capabilityOwners, ok := k.GetOwners(ctx, i)
		if !ok || len(capabilityOwners.Owners) == 0 {
			continue
		}

		genOwner := types.GenesisOwners{
			Index:  i,
			Owners: capabilityOwners,
		}
		owners = append(owners, genOwner)
	}

	return types.GenesisState{
		Index:  index,
		Owners: owners,
	}
}
