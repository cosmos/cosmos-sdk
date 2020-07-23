package v039

import (
	"fmt"

	v038auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v0_38"
)

// Migrate accepts exported x/auth genesis state from v0.38 and migrates it to
// v0.40 x/auth genesis state. The migration includes:
//
// - Removing coins from account encoding.
func Migrate(authGenState v038auth.GenesisState) v038auth.GenesisState {
	for _, account := range authGenState.Accounts {
		// set coins to nil and allow the JSON encoding to omit coins
		if err := account.SetCoins(nil); err != nil {
			panic(fmt.Sprintf("failed to set account coins to nil: %s", err))
		}
	}

	authGenState.Accounts = v038auth.SanitizeGenesisAccounts(authGenState.Accounts)
	return authGenState
}
