package bank

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

// InitGenesis initializes the bank module's state from a given genesis state.
func InitGenesis(ctx sdk.Context, keeper keeper.Keeper, genState types.GenesisState) {
	keeper.SetParams(ctx, genState.Params)

	var totalSupply sdk.Coins

	genState.Balances = types.SanitizeGenesisBalances(genState.Balances)
	for _, balance := range genState.Balances {
		if err := keeper.ValidateBalance(ctx, balance.Address); err != nil {
			panic(err)
		}

		if err := keeper.SetBalances(ctx, balance.Address, balance.Coins); err != nil {
			panic(fmt.Errorf("error on setting balances %w", err))
		}

		totalSupply = totalSupply.Add(balance.Coins...)
	}

	if genState.Supply.Empty() {
		genState.Supply = totalSupply
	}

	keeper.SetSupply(ctx, types.NewSupply(genState.Supply))
}

// ExportGenesis returns the bank module's genesis state.
func ExportGenesis(ctx sdk.Context, keeper keeper.Keeper) types.GenesisState {
	balancesSet := make(map[string]sdk.Coins)

	keeper.IterateAllBalances(ctx, func(addr sdk.AccAddress, balance sdk.Coin) bool {
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

	return types.NewGenesisState(keeper.GetParams(ctx), balances, keeper.GetSupply(ctx).GetTotal())
}

// ValidateGenesis performs basic validation of supply genesis data returning an
// error for any failed validation criteria.
func ValidateGenesis(data types.GenesisState) error {
	if err := data.Params.Validate(); err != nil {
		return err
	}
	return types.NewSupply(data.Supply).ValidateBasic()
}
