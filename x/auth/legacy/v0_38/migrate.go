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

	// TODO: migrate and convert accounts

	return NewGenesisState(authGenState.Params, accounts)
}
