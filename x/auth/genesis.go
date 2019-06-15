package auth

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// InitGenesis - Init store state from genesis data
//
// CONTRACT: old coins from the FeeCollectionKeeper need to be transferred through
// a genesis port script to the new fee collector account
func InitGenesis(ctx sdk.Context, ak AccountKeeper, data GenesisState) {
	ak.SetParams(ctx, data.Params)

	feeCollectorAcc := ak.GetAccount(ctx, types.FeeCollectorAddr)
	if feeCollectorAcc == nil {
		panic(fmt.Sprintf("fee collector module account has not been set"))
	}
}

// ExportGenesis returns a GenesisState for a given context and keeper
func ExportGenesis(ctx sdk.Context, ak AccountKeeper) GenesisState {
	params := ak.GetParams(ctx)
	return NewGenesisState(params)
}
