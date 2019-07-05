package genaccounts

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

// export genesis for all accounts
func ExportGenesis(ctx sdk.Context, accountKeeper AccountKeeper) GenesisState {

	// iterate to get the accounts
	accounts := []GenesisAccount{}
	accountKeeper.IterateAccounts(ctx,
		func(acc auth.Account) (stop bool) {
			account, err := NewGenesisAccountI(acc)
			if err != nil {
				panic(err)
			}
			accounts = append(accounts, account)
			return false
		},
	)

	return accounts
}
