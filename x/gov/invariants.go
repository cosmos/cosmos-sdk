package gov

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

// RegisterInvariants registers all governance invariants
func RegisterInvariants(ir sdk.InvariantRouter, k Keeper) {

	ir.RegisterRoute(types.ModuleName, "deposits",
		DepositsInvariant(k))
}

// AllInvariants runs all invariants of the governance module
func AllInvariants(k Keeper) sdk.Invariant {

	return func(ctx sdk.Context) error {
		return DepositsInvariant(k)(ctx)
	}
}

// DepositsInvariant checks that the module account coins reflects the deposit amounts
// held on store
func DepositsInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) error {
		// TODO: add invariant
		// - Create IterateDeposits and fix keys to use bytes instead of strings
		return nil
	}
}
