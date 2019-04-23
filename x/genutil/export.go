package genutil

import (
	"github.com/cosmos/cosmos-sdk/cmd/gaia/app"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

func ExportGenesis(accountKeeper AccountKeeper) GenesisState {

	// iterate to get the accounts
	accounts := []GenesisAccount{}
	app.accountKeeper.IterateAccounts(ctx,
		func(acc auth.Account) (stop bool) {
			account := NewGenesisAccountI(acc)
			accounts = append(accounts, account)
			return false
		},
	)

	var genesisState GenesisState
	genesisState.GenesisAccounts = accounts
	return genesisState
}
