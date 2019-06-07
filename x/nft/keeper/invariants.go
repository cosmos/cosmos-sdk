package keeper

// // RegisterInvariants registers all supply invariants
// func RegisterInvariants(ir sdk.InvariantRouter, k Keeper) {
// 	ir.RegisterRoute(
// 		types.ModuleName, "supply",
// 		SupplyInvariant(k),
// 	)
// }

// // AllInvariants runs all invariants of the nfts module.
// func AllInvariants(k Keeper) sdk.Invariant {
// 	return func(ctx sdk.Context) error {
// 		return SupplyInvariant(k)(ctx)
// 	}
// }

// // SupplyInvariant checks that the total amount of nfts on collections matches the total amount owned by addresses
// func SupplyInvariant(k Keeper) sdk.Invariant {

// 	return func(ctx sdk.Context) error {
// 		collectionsSupply := uint(0)
// 		balancesSupply := uint(0)

// 		k.IterateCollections(ctx, func(collection types.Collection) bool {
// 			collectionsSupply += collection.Supply()
// 			return false
// 		})

// 		// TODO: make this invariant per collection, not in total
// 		k.IterateBalances(ctx, BalancesKeyPrefix, func(_ sdk.AccAddress, collection types.Collection) bool {
// 			balancesSupply += collection.Supply()
// 			return false
// 		})

// 		if collectionsSupply != balancesSupply {
// 			return fmt.Errorf("total NFTs supply invariance:\n"+
// 				"\ttotal collections supply: %d\n"+
// 				"\tsum of NFT balances: %d", collectionsSupply, balancesSupply)
// 		}

// 		return nil
// 	}
// }
