package auth

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
)

// InitGenesis - Init store state from genesis data
//
// CONTRACT: old coins from the FeeCollectionKeeper need to be transferred through
// a genesis port script to the new fee collector account
func InitGenesis(ctx sdk.Context, ak AccountKeeper, data GenesisState) {
	ak.SetParams(ctx, data.Params)
	data.Accounts = SanitizeGenesisAccounts(data.Accounts)

	for _, a := range data.Accounts {
		acc := ak.NewAccount(ctx, a)
		ak.SetAccount(ctx, acc)
	}
}

// ExportGenesis returns a GenesisState for a given context and keeper
func ExportGenesis(ctx sdk.Context, ak AccountKeeper) GenesisState {
	params := ak.GetParams(ctx)

	var genAccounts exported.GenesisAccounts
	ak.IterateAccounts(ctx, func(account exported.Account) bool {
		genAccount := account.(exported.GenesisAccount)
		genAccounts = append(genAccounts, genAccount)
		return false
	})

	return NewGenesisState(params, genAccounts)
}
