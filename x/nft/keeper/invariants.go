package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/nft/types"
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
		collectionsSupply := 0
		ownersSupply := 0

		k.IterateCollections(ctx, func(collection types.Collection) bool {
			collectionsSupply += collection.Supply()
			return false
		})

		k.IterateOwners(ctx, func(owner types.Owner) bool {
			ownersSupply += owner.Supply()
			return false
		})

		if collectionsSupply != ownersSupply {
			return fmt.Errorf("total NFTs supply invariance:\n"+
				"\ttotal collections supply: %d\n"+
				"\tsum of NFTs by owner: %d", collectionsSupply, ownersSupply)
		}

		return nil
	}
}
