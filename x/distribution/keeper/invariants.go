package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/staking/exported"
)

// register all distribution invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "nonnegative-outstanding",
		NonNegativeOutstandingInvariant(k))
	ir.RegisterRoute(types.ModuleName, "can-withdraw",
		CanWithdrawInvariant(k))
	ir.RegisterRoute(types.ModuleName, "reference-count",
		ReferenceCountInvariant(k))
	ir.RegisterRoute(types.ModuleName, "module-account",
		ModuleAccountInvariant(k))
}

// AllInvariants runs all invariants of the distribution module
func AllInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) error {
		err := CanWithdrawInvariant(k)(ctx)
		if err != nil {
			return err
		}
		err = NonNegativeOutstandingInvariant(k)(ctx)
		if err != nil {
			return err
		}
		err = ReferenceCountInvariant(k)(ctx)
		if err != nil {
			return err
		}
		return ModuleAccountInvariant(k)(ctx)
	}
}

// NonNegativeOutstandingInvariant checks that outstanding unwithdrawn fees are never negative
func NonNegativeOutstandingInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) error {

		var outstanding sdk.DecCoins

		k.IterateValidatorOutstandingRewards(ctx, func(_ sdk.ValAddress, rewards types.ValidatorOutstandingRewards) (stop bool) {
			outstanding = rewards
			if outstanding.IsAnyNegative() {
				return true
			}
			return false
		})

		if outstanding.IsAnyNegative() {
			return fmt.Errorf("negative outstanding coins: %v", outstanding)
		}

		return nil

	}
}

// CanWithdrawInvariant checks that current rewards can be completely withdrawn
func CanWithdrawInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) error {

		// cache, we don't want to write changes
		ctx, _ = ctx.CacheContext()

		var remaining sdk.DecCoins

		valDelegationAddrs := make(map[string][]sdk.AccAddress)
		for _, del := range k.stakingKeeper.GetAllSDKDelegations(ctx) {
			valAddr := del.GetValidatorAddr().String()
			valDelegationAddrs[valAddr] = append(valDelegationAddrs[valAddr], del.GetDelegatorAddr())
		}

		// iterate over all validators
		k.stakingKeeper.IterateValidators(ctx, func(_ int64, val exported.ValidatorI) (stop bool) {
			_, _ = k.WithdrawValidatorCommission(ctx, val.GetOperator())

			delegationAddrs, ok := valDelegationAddrs[val.GetOperator().String()]
			if ok {
				for _, delAddr := range delegationAddrs {
					if _, err := k.WithdrawDelegationRewards(ctx, delAddr, val.GetOperator()); err != nil {
						panic(err)
					}
				}
			}

			remaining = k.GetValidatorOutstandingRewards(ctx, val.GetOperator())
			if len(remaining) > 0 && remaining[0].Amount.LT(sdk.ZeroDec()) {
				return true
			}

			return false
		})

		if len(remaining) > 0 && remaining[0].Amount.LT(sdk.ZeroDec()) {
			return fmt.Errorf("negative remaining coins: %v", remaining)
		}

		return nil
	}
}

// ReferenceCountInvariant checks that the number of historical rewards records is correct
func ReferenceCountInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) error {

		valCount := uint64(0)
		k.stakingKeeper.IterateValidators(ctx, func(_ int64, val exported.ValidatorI) (stop bool) {
			valCount++
			return false
		})
		dels := k.stakingKeeper.GetAllSDKDelegations(ctx)
		slashCount := uint64(0)
		k.IterateValidatorSlashEvents(ctx,
			func(_ sdk.ValAddress, _ uint64, _ types.ValidatorSlashEvent) (stop bool) {
				slashCount++
				return false
			})

		// one record per validator (last tracked period), one record per
		// delegation (previous period), one record per slash (previous period)
		expected := valCount + uint64(len(dels)) + slashCount
		count := k.GetValidatorHistoricalReferenceCount(ctx)

		if count != expected {
			return fmt.Errorf("unexpected number of historical rewards records: "+
				"expected %v (%v vals + %v dels + %v slashes), got %v",
				expected, valCount, len(dels), slashCount, count)
		}

		return nil
	}
}

// ModuleAccountInvariant checks that the coins held by the distr ModuleAccount
// is consistent with the sum of validator outstanding rewards
func ModuleAccountInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) error {

		var expectedCoins sdk.DecCoins
		k.IterateValidatorOutstandingRewards(ctx, func(_ sdk.ValAddress, rewards types.ValidatorOutstandingRewards) (stop bool) {
			expectedCoins = expectedCoins.Add(rewards)
			return false
		})

		communityPool := k.GetFeePoolCommunityCoins(ctx)
		expectedInt, _ := expectedCoins.Add(communityPool).TruncateDecimal()

		macc := k.GetDistributionAccount(ctx)

		if !macc.GetCoins().IsEqual(expectedInt) {
			return fmt.Errorf("distribution ModuleAccount coins invariance:\n"+
				"\texpected ModuleAccount coins: %s\n"+
				"\tdistribution ModuleAccount coins : %s", expectedInt, macc.GetCoins())
		}

		return nil
	}
}
