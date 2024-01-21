package keeper

import (
	"context"

	"cosmossdk.io/x/mint/types"
)

// InitGenesis new mint genesis
func (keeper Keeper) InitGenesis(ctx context.Context, ak types.AccountKeeper, data *types.GenesisState) {
	if err := keeper.Minter.Set(ctx, data.Minter); err != nil {
		panic(err)
	}
	ak.GetModuleAccount(ctx, types.ModuleName)
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func (keeper Keeper) ExportGenesis(ctx context.Context) *types.GenesisState {
	minter, err := keeper.Minter.Get(ctx)
	if err != nil {
		panic(err)
	}
	return types.NewGenesisState(minter)
}
