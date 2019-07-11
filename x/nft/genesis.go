package nft

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis sets nft information for genesis.
func InitGenesis(ctx sdk.Context, k Keeper, data GenesisState) {
	k.SetOwners(ctx, data.Owners)

	for _, c := range data.Collections {
		k.SetCollection(ctx, c.Denom, c)
	}

}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, k Keeper) GenesisState {
	return NewGenesisState(k.GetOwners(ctx), k.GetCollections(ctx))
}
