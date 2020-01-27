package bank

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initializes the bank module's state from a given genesis state.
func InitGenesis(ctx sdk.Context, keeper Keeper, genState GenesisState) {
	keeper.SetSendEnabled(ctx, genState.SendEnabled)

	genState.Balances = SanitizeGenesisBalances(genState.Balances)
	for _, balance := range genState.Balances {
		if err := keeper.ValidateBalance(ctx, balance.Address); err != nil {
			panic(err)
		}

		keeper.SetBalances(ctx, balance.Address, balance.Coins)
	}
}

// ExportGenesis returns the bank module's genesis state.
func ExportGenesis(ctx sdk.Context, keeper Keeper) GenesisState {
	balancesSet := make(map[string]sdk.Coins)

	keeper.IterateAllBalances(ctx, func(addr sdk.AccAddress, balance sdk.Coin) bool {
		balancesSet[addr.String()] = balancesSet[addr.String()].Add(balance)
		return false
	})

	balances := []Balance{}

	for addr, coins := range balancesSet {
		balances = append(balances, Balance{
			Address: sdk.AccAddress([]byte(addr)),
			Coins:   coins,
		})
	}

	return NewGenesisState(keeper.GetSendEnabled(ctx), balances)
}
