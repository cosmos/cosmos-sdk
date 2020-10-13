package v040

import (
	v039auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v039"
	v036supply "github.com/cosmos/cosmos-sdk/x/bank/legacy/v036"
	v038bank "github.com/cosmos/cosmos-sdk/x/bank/legacy/v038"
	v040bank "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// Migrate accepts exported v0.39 x/auth and v0.38 x/bank genesis state and
// migrates it to v0.40 x/bank genesis state. The migration includes:
//
// - Moving balances from x/auth to x/bank genesis state.
// - Moving supply from x/supply to x/bank genesis state.
// - Re-encode in v0.40 GenesisState.
func Migrate(
	bankGenState v038bank.GenesisState,
	authGenState v039auth.GenesisState,
	supplyGenState v036supply.GenesisState,
) *v040bank.GenesisState {
	balances := make([]v040bank.Balance, len(authGenState.Accounts))
	for i, acc := range authGenState.Accounts {
		balances[i] = v040bank.Balance{
			Address: acc.GetAddress().String(),
			Coins:   acc.GetCoins(),
		}
	}

	return &v040bank.GenesisState{
		Params: v040bank.Params{
			SendEnabled:        []*v040bank.SendEnabled{},
			DefaultSendEnabled: bankGenState.SendEnabled,
		},
		Balances:      balances,
		Supply:        supplyGenState.Supply,
		DenomMetadata: []v040bank.Metadata{},
	}
}
