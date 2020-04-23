package capability

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k Keeper, genState GenesisState) {
	k.SetIndex(ctx, genState.Index)

	// set owners for each index and initialize capability
	for i, owners := range genState.Owners {
		index, err := strconv.Atoi(i)
		if err != nil {
			panic(err)
		}
		k.SetOwners(ctx, uint64(index), owners)
		k.InitializeCapability(ctx, uint64(index), owners)
	}
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k Keeper) GenesisState {
	index := k.GetLatestIndex(ctx)
	owners := make(map[string]CapabilityOwners)
	for i := uint64(1); i < index; i++ {
		iOwners, ok := k.GetOwners(ctx, i)
		// Skip if nil
		if !ok {
			continue
		}
		// only export non-empty owners
		if len(iOwners.Owners) != 0 {
			owners[strconv.Itoa(int(i))] = *iOwners
		}
	}
	return GenesisState{
		Index:  index,
		Owners: owners,
	}
}
