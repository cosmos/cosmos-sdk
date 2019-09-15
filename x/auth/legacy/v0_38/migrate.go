package v038

import (
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	v036auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v0_36"
	v036genaccounts "github.com/cosmos/cosmos-sdk/x/genaccounts/legacy/v0_36"
)

// Migrate accepts exported genesis state from v0.34 and migrates it to v0.38
// genesis state.
func Migrate(authGenState v036auth.GenesisState, genAccountsGenState v036genaccounts.GenesisState) GenesisState {
	accounts := make(exported.GenesisAccounts, len(genAccountsGenState))

	for i, acc := range genAccountsGenState {
		var genAccount exported.GenesisAccount

		switch {
		case acc.StartTime != 0 && acc.EndTime != 0:
			// continuous vesting account type

		case acc.EndTime != 0:
			// delayed vesting account type

		case acc.ModuleName != "":
			// module account type

		default:
			// standard account type
		}

		accounts[i] = genAccount
	}

	return NewGenesisState(authGenState.Params, accounts)
}
