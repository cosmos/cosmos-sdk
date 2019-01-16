package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// initialize rewards for a new validator
func (k Keeper) initializeValidator(ctx sdk.Context, val sdk.Validator) {
	// set initial historical rewards (period 0)
	k.SetValidatorHistoricalRewards(ctx, val.GetOperator(), 0, types.ValidatorHistoricalRewards{})

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
		current = rewards.Rewards.QuoDec(sdk.NewDecFromInt(val.GetTokens()))
	}

	// fetch historical rewards for last period
	historical := k.GetValidatorHistoricalRewards(ctx, val.GetOperator(), rewards.Period-1)

	// fet new historical rewards
	k.SetValidatorHistoricalRewards(ctx, val.GetOperator(), rewards.Period, historical.Plus(current))

	// set current rewards, incrementing period by 1
	k.SetValidatorCurrentRewards(ctx, val.GetOperator(), types.NewValidatorCurrentRewards(sdk.DecCoins{}, rewards.Period+1))

	return rewards.Period
}

func (k Keeper) updateValidatorSlashFraction(ctx sdk.Context, valAddr sdk.ValAddress, fraction sdk.Dec) {
	height := uint64(ctx.BlockHeight())
	currentFraction := sdk.ZeroDec()
	currentPeriod := k.GetValidatorCurrentRewards(ctx, valAddr).Period
	current, found := k.GetValidatorSlashEvent(ctx, valAddr, height)
	if found {
		// there has already been a slash event this height,
		// and we don't need to store more than one,
		// so just update the current slash fraction
		currentFraction = current.Fraction
	}
	currentMultiplicand := sdk.OneDec().Sub(currentFraction)
	newMultiplicand := sdk.OneDec().Sub(fraction)
	updatedFraction := sdk.OneDec().Sub(currentMultiplicand.Mul(newMultiplicand))
	k.SetValidatorSlashEvent(ctx, valAddr, height, types.NewValidatorSlashEvent(currentPeriod, updatedFraction))
}
