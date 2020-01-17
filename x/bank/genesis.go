package bank

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis sets distribution information for genesis.
func InitGenesis(ctx sdk.Context, keeper Keeper, data GenesisState) {
	keeper.SetSendEnabled(ctx, data.SendEnabled)

	data.GenesisBalances = SanitizeGenesisBalances(data.GenesisBalances)

	for _, genBalance := range data.GenesisBalances {
		keeper.SetCoins(ctx, genBalance.Address, genBalance.Coins)
	}
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper Keeper) GenesisState {

	var genesisBalancesMap map[string]sdk.Coins

	keeper.IterateAllBalances(ctx, func(addr sdk.AccAddress, balance sdk.Coin) bool {
		genesisBalancesMap[string(addr.Bytes())] = genesisBalancesMap[string(addr.Bytes())].Add(balance)
		return false
	})

	var genesisBalances []GenesisBalance

	for addr, coins := range genesisBalancesMap {
		genesisBalances = append(genesisBalances, GenesisBalance{
			Address: sdk.AccAddress([]byte(addr)),
			Coins:   coins,
		})
	}

	return NewGenesisState(
		keeper.GetSendEnabled(ctx),
		genesisBalances,
	)
}
