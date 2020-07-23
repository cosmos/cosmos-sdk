package v040

import (
	v038auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v0_38"
	v038bank "github.com/cosmos/cosmos-sdk/x/bank/legacy/v0_38"
)

// Migrate accepts exported x/auth and x/bank genesis state from v0.38/v0.39 and
// migrates it to v0.40 x/bank genesis state. The migration includes:
//
// - Moving balances from x/auth to x/bank genesis state.
func Migrate(bankGenState v038bank.GenesisState, authGenState v038auth.GenesisState) GenesisState {
	balances := make([]Balance, len(authGenState.Accounts))
	for i, acc := range authGenState.Accounts {
		balances[i] = Balance{
			Address: acc.GetAddress(),
			Coins:   acc.GetCoins(),
		}
	}

	return NewGenesisState(bankGenState.SendEnabled, balances)
}
