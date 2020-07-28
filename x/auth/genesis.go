package auth

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// InitGenesis - Init store state from genesis data
//
// CONTRACT: old coins from the FeeCollectionKeeper need to be transferred through
// a genesis port script to the new fee collector account
func InitGenesis(ctx sdk.Context, ak keeper.AccountKeeper, data types.GenesisState) {
	ak.SetParams(ctx, data.Params)

	accounts := make([]types.GenesisAccount, len(data.Accounts))
	for i, genAcc := range data.Accounts {
		acc, ok := genAcc.GetCachedValue().(types.GenesisAccount)
		if !ok {
			panic("expected account")
		}
		accounts[i] = acc
	}
	accounts = types.SanitizeGenesisAccounts(accounts)

	for _, a := range accounts {
		acc := ak.NewAccount(ctx, a)
		ak.SetAccount(ctx, acc)
	}

	ak.GetModuleAccount(ctx, types.FeeCollectorName)
}

// ExportGenesis returns a GenesisState for a given context and keeper
func ExportGenesis(ctx sdk.Context, ak keeper.AccountKeeper) types.GenesisState {
	params := ak.GetParams(ctx)

	var genAccounts types.GenesisAccounts
	ak.IterateAccounts(ctx, func(account types.AccountI) bool {
		genAccount := account.(types.GenesisAccount)
		genAccounts = append(genAccounts, genAccount)
		return false
	})

	return types.NewGenesisState(params, genAccounts)
}
