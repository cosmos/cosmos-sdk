package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/supply/internal/types"
)

// RegisterInvariants register all supply invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "total-supply", TotalSupply(k))
}

// AllInvariants runs all invariants of the supply module.
func AllInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) error {
		return TotalSupply(k)(ctx)
	}
}

// TotalSupply checks that the total supply reflects all the coins held in accounts
func TotalSupply(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) error {
		var expectedTotal sdk.Coins
		supply := k.GetSupply(ctx)

		k.ak.IterateAccounts(ctx, func(acc exported.Account) bool {
			expectedTotal = expectedTotal.Add(acc.GetCoins())
			return false
		})

		if !expectedTotal.IsEqual(supply.Total) {
			return fmt.Errorf("total supply invariance:\n"+
				"\tsum of accounts coins: %v\n"+
				"\tsupply.Total:          %v", expectedTotal, supply.Total)
		}

		return nil
	}
}
