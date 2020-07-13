package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

// InitGenesis initializes the bank module's state from a given genesis state.
func (k BaseKeeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	k.SetParams(ctx, genState.Params)

	var totalSupply sdk.Coins

	genState.Balances = types.SanitizeGenesisBalances(genState.Balances)
	for _, balance := range genState.Balances {
		if err := k.ValidateBalance(ctx, balance.Address); err != nil {
			panic(err)
		}

		if err := k.SetBalances(ctx, balance.Address, balance.Coins); err != nil {
			panic(fmt.Errorf("error on setting balances %w", err))
		}

		totalSupply = totalSupply.Add(balance.Coins...)
	}

	if genState.Supply.Empty() {
		genState.Supply = totalSupply
	}

	k.SetSupply(ctx, types.NewSupply(genState.Supply))
}

// ExportGenesis returns the bank module's genesis state.
func (k BaseKeeper) ExportGenesis(ctx sdk.Context) types.GenesisState {
	balancesSet := make(map[string]sdk.Coins)

	k.IterateAllBalances(ctx, func(addr sdk.AccAddress, balance sdk.Coin) bool {
		balancesSet[addr.String()] = balancesSet[addr.String()].Add(balance)
		return false
	})

	balances := []types.Balance{}

	for addrStr, coins := range balancesSet {
		addr, err := sdk.AccAddressFromBech32(addrStr)
		if err != nil {
			panic(fmt.Errorf("failed to convert address from string: %w", err))
		}

		balances = append(balances, types.Balance{
			Address: addr,
			Coins:   coins,
		})
	}

	return types.NewGenesisState(k.GetParams(ctx), balances, k.GetSupply(ctx).GetTotal())
}
