package nft

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis sets distribution information for genesis.
func InitGenesis(ctx sdk.Context, k Keeper, data GenesisState) {
	for _, c := range data.Collections {
		k.SetCollection(ctx, c.Denom, c)
	}

	for _, b := range data.Balances {
		k.SetBalance(ctx, b.Owner, b.Collection)
	}
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, k Keeper) GenesisState {
	return NewGenesisState(k.GetBalances(ctx), k.GetCollections(ctx))
}
