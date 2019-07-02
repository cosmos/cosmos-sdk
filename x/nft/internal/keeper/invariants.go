package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/nft/internal/types"
)

// RegisterInvariants registers all supply invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(
		types.ModuleName, "supply",
		SupplyInvariant(k),
	)
}

// AllInvariants runs all invariants of the nfts module.
func AllInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) error {
		return SupplyInvariant(k)(ctx)
	}
}

// SupplyInvariant checks that the total amount of nfts on collections matches the total amount owned by addresses
func SupplyInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) error {
		var collectionsSupply map[string]int
		var ownersCollectionsSupply map[string]int

		k.IterateCollections(ctx, func(collection types.Collection) bool {
			collectionsSupply[collection.Denom] = collection.Supply()
			return false
		})

		k.IterateOwners(ctx, func(owner types.Owner) bool {
			for _, idCollection := range owner.IDCollections {
				ownersCollectionsSupply[idCollection.Denom] = idCollection.Supply()
			}
			return false
		})

		for denom, supply := range collectionsSupply {
			if supply != ownersCollectionsSupply[denom] {
				return fmt.Errorf("total NFTs supply invariance:\n"+
					"\ttotal %s NFTs supply: %d\n"+
					"\tsum of %s NFTs by owner: %d", denom, supply, denom, ownersCollectionsSupply[denom])
			}
		}
		return nil
	}
}
