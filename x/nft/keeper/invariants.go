package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/nft/types"
)

// RegisterInvariants registers all supply invariants
func RegisterInvariants(ck CrisisKeeper, k Keeper) {
	ck.RegisterRoute(
		ModuleName, "supply",
		SupplyInvariant(k),
	)
}

// AllInvariants runs all invariants of the nfts module.
func AllInvariants(k Keeper) sdk.Invariant {

	return func(ctx sdk.Context) error {
		err := SupplyInvariant(k)(ctx)
		if err != nil {
			return err
		}

		// err := OwnershipInvariant(k)(ctx)
		// if err != nil {
		// 	return err
		// }

		return nil
	}
}

// SupplyInvariant checks that the total amount of nfts on collections matches the total amount owned by addresses
func SupplyInvariant(k Keeper) sdk.Invariant {

	return func(ctx sdk.Context) error {
		collectionsSupply := uint(0)
		balancesSupply := uint(0)

		k.IterateCollections(ctx, func(collection types.Collection) bool {
			collectionsSupply += collection.Supply()
			return false
		})

		// TODO: make this invariant per collection, not in total
		k.IterateBalances(ctx, func(_ sdk.AccAddress, collection types.Collection) bool {
			balancesSupply += collection.Supply()
			return false
		})

		if collectionsSupply != balancesSupply {
			return fmt.Errorf("total NFTs supply invariance:\n"+
				"\ttotal collections supply: %d\n"+
				"\tsum of NFT balances: %d", collectionsSupply, balancesSupply)
		}

		return nil
	}
}
