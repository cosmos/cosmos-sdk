package changepubkey

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/changepubkey/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/changepubkey/types"
)

// InitGenesis - Init store state from genesis data
func InitGenesis(ctx sdk.Context, pk keeper.ChangePubKeyKeeper, data types.GenesisState) {
	pk.SetParams(ctx, data.Params)
}

// ExportGenesis returns a GenesisState for a given context and keeper
func ExportGenesis(ctx sdk.Context, pk keeper.ChangePubKeyKeeper) *types.GenesisState {
	params := pk.GetParams(ctx)
	return types.NewGenesisState(params)
}
