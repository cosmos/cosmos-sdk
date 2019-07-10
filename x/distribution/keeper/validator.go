package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/staking/exported"
)

// initialize rewards for a new validator
func (k Keeper) initializeValidator(ctx sdk.Context, val exported.ValidatorI) {
	// set initial historical rewards (period 0) with reference count of 1
	k.SetValidatorHistoricalRewards(ctx, val.GetOperator(), 0, types.NewValidatorHistoricalRewards(sdk.DecCoins{}, 1))

	// set current rewards (starting at period 1)
	k.SetValidatorCurrentRewards(ctx, val.GetOperator(), types.NewValidatorCurrentRewards(sdk.DecCoins{}, 1))

	// set accumulated commission
	k.SetValidatorAccumulatedCommission(ctx, val.GetOperator(), types.InitialValidatorAccumulatedCommission())

	// set outstanding rewards
	k.SetValidatorOutstandingRewards(ctx, val.GetOperator(), sdk.DecCoins{})
}

// increment validator period, returning the period just ended
func (k Keeper) incrementValidatorPeriod(ctx sdk.Context, val exported.ValidatorI) uint64 {
	// fetch current rewards
	rewards := k.GetValidatorCurrentRewards(ctx, val.GetOperator())

	// calculate current ratio
	var current sdk.DecCoins
	if val.GetTokens().IsZero() {

		// can't calculate ratio for zero-token validators
		// ergo we instead add to the community pool
		feePool := k.GetFeePool(ctx)
		outstanding := k.GetValidatorOutstandingRewards(ctx, val.GetOperator())
		feePool.CommunityPool = feePool.CommunityPool.Add(rewards.Rewards)
		outstanding = outstanding.Sub(rewards.Rewards)
		k.SetFeePool(ctx, feePool)
		k.SetValidatorOutstandingRewards(ctx, val.GetOperator(), outstanding)

		current = sdk.DecCoins{}
	} else {
		// note: necessary to truncate so we don't allow withdrawing more rewards than owed
		current = rewards.Rewards.QuoDecTruncate(val.GetTokens().ToDec())
	}

	// fetch historical rewards for last period
	historical := k.GetValidatorHistoricalRewards(ctx, val.GetOperator(), rewards.Period-1).CumulativeRewardRatio

	// decrement reference count
	k.decrementReferenceCount(ctx, val.GetOperator(), rewards.Period-1)

	// set new historical rewards with reference count of 1
	k.SetValidatorHistoricalRewards(ctx, val.GetOperator(), rewards.Period, types.NewValidatorHistoricalRewards(historical.Add(current), 1))

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
	if fraction.GT(sdk.OneDec()) || fraction.LT(sdk.ZeroDec()) {
		panic(fmt.Sprintf("fraction must be >=0 and <=1, current fraction: %v", fraction))
	}

	val := k.stakingKeeper.Validator(ctx, valAddr)

	// increment current period
	newPeriod := k.incrementValidatorPeriod(ctx, val)

	// increment reference count on period we need to track
	k.incrementReferenceCount(ctx, valAddr, newPeriod)

	slashEvent := types.NewValidatorSlashEvent(newPeriod, fraction)
	height := uint64(ctx.BlockHeight())

	k.SetValidatorSlashEvent(ctx, valAddr, height, newPeriod, slashEvent)
}
