package keeper

import (
	"context"

	"cosmossdk.io/x/epochs/types"
)

// InitGenesis sets epoch info from genesis
func (k Keeper) InitGenesis(ctx context.Context, genState types.GenesisState) {
	for _, epoch := range genState.Epochs {
		err := k.AddEpochInfo(ctx, epoch)
		if err != nil {
			panic(err)
		}
	}
}

// ExportGenesis returns the capability module's exported genesis.
func (k Keeper) ExportGenesis(ctx context.Context) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Epochs = k.AllEpochInfos(ctx)
	return genesis
}
