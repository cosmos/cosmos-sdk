package simulation

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/mock/simulation"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

// AllInvariants runs all invariants of the distribution module
func AllInvariants(d distr.Keeper, stk staking.Keeper) simulation.Invariant {
	return func(ctx sdk.Context) error {
		err := CanWithdrawInvariant(d, stk)(ctx)
		if err != nil {
			return err
		}
		err = NonNegativeOutstandingInvariant(d)(ctx)
		if err != nil {
			return err
		}
		err = ReferenceCountInvariant(d, stk)(ctx)
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
			return fmt.Errorf("negative outstanding coins: %v", outstanding)
		}
		return nil
	}
}

// CanWithdrawInvariant checks that current rewards can be completely withdrawn
func CanWithdrawInvariant(k distr.Keeper, sk staking.Keeper) simulation.Invariant {
	return func(ctx sdk.Context) error {

		// cache, we don't want to write changes
		ctx, _ = ctx.CacheContext()

		// iterate over all bonded validators, withdraw commission
		sk.IterateValidators(ctx, func(_ int64, val sdk.Validator) (stop bool) {
			_ = k.WithdrawValidatorCommission(ctx, val.GetOperator())
			return false
		})

		// iterate over all current delegations, withdraw rewards
		dels := sk.GetAllDelegations(ctx)
		for _, delegation := range dels {
			_ = k.WithdrawDelegationRewards(ctx, delegation.DelegatorAddr, delegation.ValidatorAddr)
		}

		remaining := k.GetOutstandingRewards(ctx)

		if len(remaining) > 0 && remaining[0].Amount.LT(sdk.ZeroDec()) {
			return fmt.Errorf("negative remaining coins: %v", remaining)
		}

		return nil
	}
}

// ReferenceCountInvariant checks that the number of historical rewards records is correct
func ReferenceCountInvariant(k distr.Keeper, sk staking.Keeper) simulation.Invariant {
	return func(ctx sdk.Context) error {

		valCount := uint64(0)
		sk.IterateValidators(ctx, func(_ int64, val sdk.Validator) (stop bool) {
			valCount++
			return false
		})
		dels := sk.GetAllDelegations(ctx)
		slashCount := uint64(0)
		k.IterateValidatorSlashEvents(ctx, func(_ sdk.ValAddress, _ uint64, _ types.ValidatorSlashEvent) (stop bool) {
			slashCount++
			return false
		})

		// one record per validator (zeroeth period), one record per delegation (previous period), one record per slash (previous period)
		expected := valCount + uint64(len(dels)) + slashCount

		count := k.GetValidatorHistoricalRewardCount(ctx)
		if count != expected {
			return fmt.Errorf("unexpected number of historical rewards records: expected %v, got %v", expected, count)
		}

		return nil
	}
}
