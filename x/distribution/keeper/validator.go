package keeper

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// initialize rewards for a new validator
func (k Keeper) initializeValidator(ctx context.Context, val stakingtypes.ValidatorI) error {
	// set initial historical rewards (period 0) with reference count of 1
	err := k.ValidatorHistoricalRewards.Set(ctx, collections.Join(val.GetOperator(), uint64(0)), types.NewValidatorHistoricalRewards(sdk.DecCoins{}, 1))
	if err != nil {
		return err
	}

	// set current rewards (starting at period 1)
	err = k.ValidatorCurrentRewards.Set(ctx, val.GetOperator(), types.NewValidatorCurrentRewards(sdk.DecCoins{}, 1))
	if err != nil {
		return err
	}

	// set accumulated commission
	err = k.ValidatorsAccumulatedCommission.Set(ctx, val.GetOperator(), types.InitialValidatorAccumulatedCommission())
	if err != nil {
		return err
	}

	// set outstanding rewards
	err = k.ValidatorOutstandingRewards.Set(ctx, val.GetOperator(), types.ValidatorOutstandingRewards{Rewards: sdk.DecCoins{}})
	return err
}

// increment validator period, returning the period just ended
func (k Keeper) IncrementValidatorPeriod(ctx context.Context, val stakingtypes.ValidatorI) (uint64, error) {
	// fetch current rewards
	rewards, err := k.ValidatorCurrentRewards.Get(ctx, val.GetOperator())
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		return 0, err
	}

	// calculate current ratio
	var current sdk.DecCoins
	if val.GetTokens().IsZero() {

		// can't calculate ratio for zero-token validators
		// ergo we instead add to the community pool
		feePool, err := k.FeePool.Get(ctx)
		if err != nil {
			return 0, err
		}

		outstanding, err := k.ValidatorOutstandingRewards.Get(ctx, val.GetOperator())
		if err != nil && !errors.Is(err, collections.ErrNotFound) {
			return 0, err
		}

		feePool.CommunityPool = feePool.CommunityPool.Add(rewards.Rewards...)
		outstanding.Rewards = outstanding.GetRewards().Sub(rewards.Rewards)
		err = k.FeePool.Set(ctx, feePool)
		if err != nil {
			return 0, err
		}

		err = k.ValidatorOutstandingRewards.Set(ctx, val.GetOperator(), outstanding)
		if err != nil {
			return 0, err
		}

		current = sdk.DecCoins{}
	} else {
		// note: necessary to truncate so we don't allow withdrawing more rewards than owed
		current = rewards.Rewards.QuoDecTruncate(math.LegacyNewDecFromInt(val.GetTokens()))
	}

	// fetch historical rewards for last period
	historical, err := k.ValidatorHistoricalRewards.Get(ctx, collections.Join(val.GetOperator(), rewards.Period-1))
	if err != nil {
		return 0, err
	}

	cumRewardRatio := historical.CumulativeRewardRatio

	// decrement reference count
	err = k.decrementReferenceCount(ctx, val.GetOperator(), rewards.Period-1)
	if err != nil {
		return 0, err
	}

	// set new historical rewards with reference count of 1
	err = k.ValidatorHistoricalRewards.Set(ctx, collections.Join(val.GetOperator(), rewards.Period), types.NewValidatorHistoricalRewards(cumRewardRatio.Add(current...), 1))
	if err != nil {
		return 0, err
	}

	// set current rewards, incrementing period by 1
	err = k.ValidatorCurrentRewards.Set(ctx, val.GetOperator(), types.NewValidatorCurrentRewards(sdk.DecCoins{}, rewards.Period+1))
	if err != nil {
		return 0, err
	}

	return rewards.Period, nil
}

// increment the reference count for a historical rewards value
func (k Keeper) incrementReferenceCount(ctx context.Context, valAddr sdk.ValAddress, period uint64) error {
	historical, err := k.ValidatorHistoricalRewards.Get(ctx, collections.Join(valAddr, period))
	if err != nil {
		return err
	}
	if historical.ReferenceCount > 2 {
		panic("reference count should never exceed 2")
	}
	historical.ReferenceCount++
	return k.ValidatorHistoricalRewards.Set(ctx, collections.Join(valAddr, period), historical)
}

// decrement the reference count for a historical rewards value, and delete if zero references remain
func (k Keeper) decrementReferenceCount(ctx context.Context, valAddr sdk.ValAddress, period uint64) error {
	historical, err := k.ValidatorHistoricalRewards.Get(ctx, collections.Join(valAddr, period))
	if err != nil {
		return err
	}

	if historical.ReferenceCount == 0 {
		panic("cannot set negative reference count")
	}
	historical.ReferenceCount--
	if historical.ReferenceCount == 0 {
		return k.ValidatorHistoricalRewards.Remove(ctx, collections.Join(valAddr, period))
	}

	return k.ValidatorHistoricalRewards.Set(ctx, collections.Join(valAddr, period), historical)
}

func (k Keeper) updateValidatorSlashFraction(ctx context.Context, valAddr sdk.ValAddress, fraction math.LegacyDec) error {
	if fraction.GT(math.LegacyOneDec()) || fraction.IsNegative() {
		panic(fmt.Sprintf("fraction must be >=0 and <=1, current fraction: %v", fraction))
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	val, err := k.stakingKeeper.Validator(ctx, valAddr)
	if err != nil {
		return err
	}

	// increment current period
	newPeriod, err := k.IncrementValidatorPeriod(ctx, val)
	if err != nil {
		return err
	}

	// increment reference count on period we need to track
	err = k.incrementReferenceCount(ctx, valAddr, newPeriod)
	if err != nil {
		return err
	}

	slashEvent := types.NewValidatorSlashEvent(newPeriod, fraction)
	height := uint64(sdkCtx.BlockHeight())

	return k.SetValidatorSlashEvent(ctx, valAddr, height, newPeriod, slashEvent)
}
