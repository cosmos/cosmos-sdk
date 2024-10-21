package keeper

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mint/types"
)

// InitGenesis new mint genesis
func (keeper Keeper) InitGenesis(ctx sdk.Context, ak types.AccountKeeper, data *types.GenesisState) {
	fmt.Println("initializing minter from genesis", data.Minter)
	if err := keeper.Minter.Set(ctx, data.Minter); err != nil {
		panic(err)
	}

	fmt.Println("initializing params from genesis", data.Params)

	if err := keeper.Params.Set(ctx, data.Params); err != nil {
		panic(err)
	}

	fmt.Println("checking after initialization")
	fmt.Println(keeper.Minter.Get(ctx))
	fmt.Println("minter is ok")
	fmt.Println(keeper.Params.Get(ctx))
	fmt.Println("params is ok")

	ak.GetModuleAccount(ctx, types.ModuleName)
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func (keeper Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	minter, err := keeper.Minter.Get(ctx)
	if err != nil {
		panic(err)
	}

	params, err := keeper.Params.Get(ctx)
	if err != nil {
		panic(err)
	}

	return types.NewGenesisState(minter, params)
}
