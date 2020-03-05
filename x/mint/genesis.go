package mint

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mint/types"
)

// initGenesis new mint genesis
func initGenesis(ctx sdk.Context, keeper Keeper, supplyKeeper types.SupplyKeeper, data GenesisState) {
	keeper.SetMinter(ctx, data.Minter)
	keeper.SetParams(ctx, data.Params)

	supplyKeeper.GetModuleAccount(ctx, ModuleName)
}

// exportGenesis returns a GenesisState for a given context and keeper.
func exportGenesis(ctx sdk.Context, keeper Keeper) GenesisState {
	minter := keeper.GetMinter(ctx)
	params := keeper.GetParams(ctx)
	return NewGenesisState(minter, params)
}
