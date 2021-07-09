package v043

import (
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

func migrateBalances(oldBalances []types.Balance) []types.Balance {
	var balances []types.Balance

	for _, b := range oldBalances {
		if !b.Coins.IsZero() {
			balances = append(balances, b)
		}
	}

	return balances
}

func MigrateJSON(oldState *types.GenesisState) *types.GenesisState {
	return &types.GenesisState{
		Params:        oldState.Params,
		Balances:      migrateBalances(oldState.Balances),
		Supply:        oldState.Supply,
		DenomMetadata: oldState.DenomMetadata,
	}
}
