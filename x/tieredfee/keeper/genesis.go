package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/tieredfee/types"
)

// InitGenesis initializes the tieredfee module's state from a given genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState *types.GenesisState) {
	k.SetParams(ctx, genState.Params)
	if genState.ParentGasUsed > 0 {
		k.SetBlockGasUsed(ctx, genState.ParentGasUsed)
	}
	for i, protoCoins := range genState.GasPrices {
		k.SetGasPrice(ctx, uint32(i), protoCoins.Coins)
	}
}

// ExportGenesis returns the tieredfee module's genesis state.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	gasPrices := k.GetAllGasPrice(ctx)
	gasUsed, found := k.GetBlockGasUsed(ctx)
	if !found {
		gasUsed = 0
	}
	return &types.GenesisState{
		Params:        k.GetParams(ctx),
		ParentGasUsed: gasUsed,
		GasPrices:     types.ToProtoGasPrices(gasPrices),
	}
}
