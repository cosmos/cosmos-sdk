package v040

import (
	"fmt"

	v039auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v0_39"
)

// Migrate accepts exported x/auth genesis state from v0.38/v0.39 and migrates
// it to v0.40 x/auth genesis state. The migration includes:
//
// - Removing coins from account encoding.
func Migrate(authGenState v039auth.GenesisState) v039auth.GenesisState {
	for _, account := range authGenState.Accounts {
		// set coins to nil and allow the JSON encoding to omit coins
		if err := account.SetCoins(nil); err != nil {
			panic(fmt.Sprintf("failed to set account coins to nil: %s", err))
		}
	}

	authGenState.Accounts = v039auth.SanitizeGenesisAccounts(authGenState.Accounts)
	return authGenState
}
