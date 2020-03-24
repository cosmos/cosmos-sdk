package bank

import (
	"fmt"

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

		if err := keeper.SetBalances(ctx, balance.Address, balance.Coins); err != nil {
			panic(fmt.Errorf("error on setting balances %w", err))
		}
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

	for addrStr, coins := range balancesSet {
		addr, err := sdk.AccAddressFromBech32(addrStr)
		if err != nil {
			panic(fmt.Errorf("failed to convert address from string: %w", err))
		}

		balances = append(balances, Balance{
			Address: addr,
			Coins:   coins,
		})
	}

	return NewGenesisState(keeper.GetSendEnabled(ctx), balances)
}
