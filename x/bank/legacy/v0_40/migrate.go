package v040

import (
	v039auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v0_39"
	v038bank "github.com/cosmos/cosmos-sdk/x/bank/legacy/v0_38"
)

// Migrate accepts exported v0.39 x/auth and v0.38 x/bank genesis state and
// migrates it to v0.40 x/bank genesis state. The migration includes:
//
// - Moving balances from x/auth to x/bank genesis state.
func Migrate(bankGenState v038bank.GenesisState, authGenState v039auth.GenesisState) GenesisState {
	balances := make([]Balance, len(authGenState.Accounts))
	for i, acc := range authGenState.Accounts {
		balances[i] = Balance{
			Address: acc.GetAddress(),
			Coins:   acc.GetCoins(),
		}
	}

	return NewGenesisState(bankGenState.SendEnabled, balances)
}
