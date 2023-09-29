package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/x/crisis/types"
)

// new crisis genesis
func (k *Keeper) InitGenesis(ctx context.Context, data *types.GenesisState) {
	if err := k.ConstantFee.Set(ctx, data.ConstantFee); err != nil {
		panic(err)
	}
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func (k *Keeper) ExportGenesis(ctx context.Context) *types.GenesisState {
	constantFee, err := k.ConstantFee.Get(ctx)
	if err != nil {
		panic(err)
	}
	return types.NewGenesisState(constantFee)
}
