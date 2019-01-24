package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// initialize rewards for a new validator
func (k Keeper) initializeValidator(ctx sdk.Context, val sdk.Validator) {
	// set initial historical rewards (period 0) with reference count of 1
	k.SetValidatorHistoricalRewards(ctx, val.GetOperator(), 0, types.NewValidatorHistoricalRewards(sdk.DecCoins{}, 1))

	// set current rewards (starting at period 1)
	k.SetValidatorCurrentRewards(ctx, val.GetOperator(), types.NewValidatorCurrentRewards(sdk.DecCoins{}, 1))

	// set accumulated commission
	k.SetValidatorAccumulatedCommission(ctx, val.GetOperator(), types.InitialValidatorAccumulatedCommission())
}

// increment validator period, returning the period just ended
func (k Keeper) incrementValidatorPeriod(ctx sdk.Context, val sdk.Validator) uint64 {
	// fetch current rewards
	rewards := k.GetValidatorCurrentRewards(ctx, val.GetOperator())

	// calculate current ratio
	var current sdk.DecCoins
	if val.GetTokens().IsZero() {

		// can't calculate ratio for zero-token validators
		// ergo we instead add to the community pool
		feePool := k.GetFeePool(ctx)
		outstanding := k.GetOutstandingRewards(ctx)
		feePool.CommunityPool = feePool.CommunityPool.Plus(rewards.Rewards)
		outstanding = outstanding.Minus(rewards.Rewards)
		k.SetFeePool(ctx, feePool)
		k.SetOutstandingRewards(ctx, outstanding)

		current = sdk.DecCoins{}
	} else {
		// note: necessary to truncate so we don't allow withdrawing more rewards than owed
		current = rewards.Rewards.QuoDecTruncate(sdk.NewDecFromInt(val.GetTokens()))
	}

	// fetch historical rewards for last period
	historical := k.GetValidatorHistoricalRewards(ctx, val.GetOperator(), rewards.Period-1).CumulativeRewardRatio

	// decrement reference count
	k.decrementReferenceCount(ctx, val.GetOperator(), rewards.Period-1)

	// set new historical rewards with reference count of 1
	k.SetValidatorHistoricalRewards(ctx, val.GetOperator(), rewards.Period, types.NewValidatorHistoricalRewards(historical.Plus(current), 1))

	// set current rewards, incrementing period by 1
	k.SetValidatorCurrentRewards(ctx, val.GetOperator(), types.NewValidatorCurrentRewards(sdk.DecCoins{}, rewards.Period+1))

	return rewards.Period
}

// increment the reference count for a historical rewards value
func (k Keeper) incrementReferenceCount(ctx sdk.Context, valAddr sdk.ValAddress, period uint64) {
	historical := k.GetValidatorHistoricalRewards(ctx, valAddr, period)
	if historical.ReferenceCount > 2 {
		panic("reference count should never exceed 2")
	}
	historical.ReferenceCount++
	k.SetValidatorHistoricalRewards(ctx, valAddr, period, historical)
}

// decrement the reference count for a historical rewards value, and delete if zero references remain
func (k Keeper) decrementReferenceCount(ctx sdk.Context, valAddr sdk.ValAddress, period uint64) {
	historical := k.GetValidatorHistoricalRewards(ctx, valAddr, period)
	if historical.ReferenceCount == 0 {
		panic("cannot set negative reference count")
	}
	historical.ReferenceCount--
	if historical.ReferenceCount == 0 {
		k.DeleteValidatorHistoricalReward(ctx, valAddr, period)
	} else {
		k.SetValidatorHistoricalRewards(ctx, valAddr, period, historical)
	}
}

func (k Keeper) updateValidatorSlashFraction(ctx sdk.Context, valAddr sdk.ValAddress, fraction sdk.Dec) {
	if fraction.GT(sdk.OneDec()) {
		panic("fraction greater than one")
	}
	height := uint64(ctx.BlockHeight())
	currentFraction := sdk.ZeroDec()
	endedPeriod := k.GetValidatorCurrentRewards(ctx, valAddr).Period - 1
	current, found := k.GetValidatorSlashEvent(ctx, valAddr, height)
	if found {
		// there has already been a slash event this height,
		// and we don't need to store more than one,
		// so just update the current slash fraction
		currentFraction = current.Fraction
	} else {
		val := k.stakingKeeper.Validator(ctx, valAddr)
		// increment current period
		endedPeriod = k.incrementValidatorPeriod(ctx, val)
		// increment reference count on period we need to track
		k.incrementReferenceCount(ctx, valAddr, endedPeriod)
	}
	currentMultiplicand := sdk.OneDec().Sub(currentFraction)
	newMultiplicand := sdk.OneDec().Sub(fraction)
	updatedFraction := sdk.OneDec().Sub(currentMultiplicand.Mul(newMultiplicand))
	if updatedFraction.LT(sdk.ZeroDec()) {
		panic("negative slash fraction")
	}
	k.SetValidatorSlashEvent(ctx, valAddr, height, types.NewValidatorSlashEvent(endedPeriod, updatedFraction))
}
