package auth

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// InitGenesis - Init store state from genesis data
func InitGenesis(ctx sdk.Context, ak AccountKeeper, fck types.FeeCollectionKeeper, data types.GenesisState) {
	ak.SetParams(ctx, data.Params)
	fck.AddCollectedFees(ctx, data.CollectedFees)
}

// ExportGenesis returns a GenesisState for a given context and keeper
func ExportGenesis(ctx sdk.Context, ak AccountKeeper, fck types.FeeCollectionKeeper) types.GenesisState {
	collectedFees := fck.GetCollectedFees(ctx)
	params := ak.GetParams(ctx)

	return types.NewGenesisState(collectedFees, params)
}
