package keeper

import (
	types2 "github.com/cosmos/cosmos-sdk/contrib/x/crisis/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// new crisis genesis
func (k *Keeper) InitGenesis(ctx sdk.Context, data *types2.GenesisState) {
	if err := k.ConstantFee.Set(ctx, data.ConstantFee); err != nil {
		panic(err)
	}
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func (k *Keeper) ExportGenesis(ctx sdk.Context) *types2.GenesisState {
	constantFee, err := k.ConstantFee.Get(ctx)
	if err != nil {
		panic(err)
	}
	return types2.NewGenesisState(constantFee)
}
