package crisis

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// new crisis genesis
func InitGenesis(ctx sdk.Context, keeper Keeper, data GenesisState) {
	keeper.SetConstantFee(ctx, data.ConstantFee)
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper Keeper) GenesisState {
	constantFee := keeper.GetConstantFee(ctx)
	return NewGenesisState(constantFee)
}
