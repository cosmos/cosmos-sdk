package nfts

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis sets distribution information for genesis.
func InitGenesis(ctx sdk.Context, k Keeper, data GenesisState) {
	for _, collection := range data.Collections {
		k.SetCollection(ctx, collection.Denom, collection)
	}

	for _, balance := range data.Balance {
		k.SetOwnerBalance(ctx, balance.Owner, balance)
	}
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, k Keeper) GenesisState {
	return NewGenesisState(k.GetCollections(ctx), k.GetOwners(ctx))
}
