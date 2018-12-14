package simulation

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/mock/simulation"
	"github.com/cosmos/cosmos-sdk/x/stake"
)

// AllInvariants runs all invariants of the distribution module
func AllInvariants(d distr.Keeper, stk stake.Keeper) simulation.Invariant {
	return func(ctx sdk.Context) error {
		err := NonNegativeOutstandingInvariant(d)(ctx)
		if err != nil {
			return err
		}
		err = CanWithdrawInvariant(d, stk)(ctx)
		if err != nil {
			return err
		}
		return nil
	}
}

// NonNegativeOutstandingInvariant checks that outstanding unwithdrawn fees are never negative
func NonNegativeOutstandingInvariant(k distr.Keeper) simulation.Invariant {
	return func(ctx sdk.Context) error {
		outstanding := k.GetOutstandingRewards(ctx)
		if outstanding.HasNegative() {
			return fmt.Errorf("Negative outstanding coins: %v", outstanding)
		}
		return nil
	}
}

// CanWithdrawInvariant checks that current rewards can be completely withdrawn
func CanWithdrawInvariant(k distr.Keeper, sk stake.Keeper) simulation.Invariant {
	return func(ctx sdk.Context) error {

		// TODO

		// all ok
		return nil
	}
}
