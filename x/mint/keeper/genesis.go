package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mint/types"
)

// InitGenesis new mint genesis
func (keeper Keeper) InitGenesis(ctx sdk.Context, ak types.AccountKeeper, data *types.GenesisState) {
	keeper.SetMinter(ctx, data.Minter)

	if err := keeper.SetParams(ctx, data.Params); err != nil {
		panic(err)
	}

	ak.GetModuleAccount(ctx, types.ModuleName)
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func (keeper Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	minter, err := keeper.GetMinter(ctx)
	if err != nil {
		panic(err)
	}

	params, err := keeper.GetParams(ctx)
	if err != nil {
		panic(err)
	}

	return types.NewGenesisState(minter, params)
}
