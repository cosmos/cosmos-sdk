package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// initialize starting info for a new delegation
func (k Keeper) initializeDelegation(ctx sdk.Context, val sdk.ValAddress, del sdk.AccAddress) {
	// period has already been incremented - we want to store the period ended by this delegation action
	previousPeriod := k.GetValidatorCurrentRewards(ctx, val).Period - 1

	// increment reference count for the period we're going to track
	k.incrementReferenceCount(ctx, val, previousPeriod)

	validator := k.stakingKeeper.Validator(ctx, val)
	delegation := k.stakingKeeper.Delegation(ctx, del, val)

	// calculate delegation stake in tokens
	// we don't store directly, so multiply delegation shares * (tokens per share)
	// note: necessary to truncate so we don't allow withdrawing more rewards than owed
	stake := delegation.GetShares().MulTruncate(validator.GetDelegatorShareExRateTruncated())
	k.SetDelegatorStartingInfo(ctx, val, del, types.NewDelegatorStartingInfo(previousPeriod, stake, uint64(ctx.BlockHeight())))
}

// calculate the rewards accrued by a delegation between two periods
func (k Keeper) calculateDelegationRewardsBetween(ctx sdk.Context, val sdk.Validator,
	startingPeriod, endingPeriod uint64, stake sdk.Dec) (rewards sdk.DecCoins) {
	// sanity check
	if startingPeriod > endingPeriod {
		panic("startingPeriod cannot be greater than endingPeriod")
	}

	// sanity check
	if stake.LT(sdk.ZeroDec()) {
		panic("stake should not be negative")
	}

	// return staking * (ending - starting)
	starting := k.GetValidatorHistoricalRewards(ctx, val.GetOperator(), startingPeriod)
	ending := k.GetValidatorHistoricalRewards(ctx, val.GetOperator(), endingPeriod)
	difference := ending.CumulativeRewardRatio.Sub(starting.CumulativeRewardRatio)
	if difference.IsAnyNegative() {
		panic("negative rewards should not be possible")
	}
	// note: necessary to truncate so we don't allow withdrawing more rewards than owed
	rewards = difference.MulDecTruncate(stake)
	return
}

// calculate the total rewards accrued by a delegation
func (k Keeper) calculateDelegationRewards(ctx sdk.Context, val sdk.Validator, del sdk.Delegation, endingPeriod uint64) (rewards sdk.DecCoins) {
	// fetch starting info for delegation
	startingInfo := k.GetDelegatorStartingInfo(ctx, del.GetValidatorAddr(), del.GetDelegatorAddr())

	if startingInfo.Height == uint64(ctx.BlockHeight()) {
		// started this height, no rewards yet
		return
	}

	startingPeriod := startingInfo.PreviousPeriod
	stake := startingInfo.Stake

	// due to rounding differences due to order of operations, stake calculated
	// with the F1 distribution mechanism  may be slightly inflated thus we
	// introduce a correction factor in order to adjust for this inflation.
	currentStake := del.GetShares().Mul(val.GetDelegatorShareExRate())
	currentStakeF1 := val.GetTokens().ToDec().MulTruncate(del.GetShares()).QuoTruncate(val.GetDelegatorShares())
	var roundingCorrectionFactor sdk.Dec
	if currentStakeF1.IsZero() {
		roundingCorrectionFactor = sdk.OneDec()
	} else {
		roundingCorrectionFactor = currentStake.QuoTruncate(currentStakeF1)
	}

	// iterate through slashes and withdraw with calculated staking for sub-intervals
	// these offsets are dependent on *when* slashes happen - namely, in BeginBlock, after rewards are allocated...
	// slashes which happened in the first block would have been before this delegation existed,
	// UNLESS they were slashes of a redelegation to this validator which was itself slashed
	// (from a fault committed by the redelegation source validator) earlier in the same BeginBlock
	startingHeight := startingInfo.Height
	// slashes this block happened after reward allocation, but we have to account for them for the stake sanity check below
	endingHeight := uint64(ctx.BlockHeight())
	if endingHeight > startingHeight {
		k.IterateValidatorSlashEventsBetween(ctx, del.GetValidatorAddr(), startingHeight, endingHeight,
			func(height uint64, event types.ValidatorSlashEvent) (stop bool) {
				endingPeriod := event.ValidatorPeriod
				if endingPeriod > startingPeriod {
					rewards = rewards.Add(k.calculateDelegationRewardsBetween(ctx, val, startingPeriod, endingPeriod, stake))
					// note: necessary to truncate so we don't allow withdrawing more rewards than owed
					stake = stake.MulTruncate(sdk.OneDec().Sub(event.Fraction)).MulTruncate(roundingCorrectionFactor)
					startingPeriod = endingPeriod
				}
				return false
			},
		)
	}

	// a stake sanity check - recalculated final stake should be less than or equal to current stake
	// here we cannot use Equals because stake is truncated when multiplied by slash fractions
	// we could only use equals if we had arbitrary-precision rationals
	if stake.GT(currentStake) {
		fmt.Printf("debug currentStake:     %v\n", currentStake)
		fmt.Printf("debug currentStakeF1:   %v\n", currentStakeF1)
		fmt.Printf("debug stake:            %v\n", stake)
		panic(fmt.Sprintf("calculated final stake for delegator %s greater than current stake: %s, %s",
			del.GetDelegatorAddr(), stake, del.GetShares().Mul(val.GetDelegatorShareExRate())))
	}

	// calculate rewards for final period
	rewards = rewards.Add(k.calculateDelegationRewardsBetween(ctx, val, startingPeriod, endingPeriod, stake))

	return rewards
}

func (k Keeper) withdrawDelegationRewards(ctx sdk.Context, val sdk.Validator, del sdk.Delegation) sdk.Error {

	// check existence of delegator starting info
	if !k.HasDelegatorStartingInfo(ctx, del.GetValidatorAddr(), del.GetDelegatorAddr()) {
		return types.ErrNoDelegationDistInfo(k.codespace)
	}

	// end current period and calculate rewards
	endingPeriod := k.incrementValidatorPeriod(ctx, val)
	rewards := k.calculateDelegationRewards(ctx, val, del, endingPeriod)
	outstanding := k.GetValidatorOutstandingRewards(ctx, del.GetValidatorAddr())

	// defensive edge case may happen on the very final digits
	// of the decCoins due to operation order of the distribution mechanism.
	// TODO log if rewards is reduced in this step
	rewards = sdk.MinSet(rewards, outstanding)

	// decrement reference count of starting period
	startingInfo := k.GetDelegatorStartingInfo(ctx, del.GetValidatorAddr(), del.GetDelegatorAddr())
	startingPeriod := startingInfo.PreviousPeriod
	k.decrementReferenceCount(ctx, del.GetValidatorAddr(), startingPeriod)

	// truncate coins, return remainder to community pool
	coins, remainder := rewards.TruncateDecimal()

	k.SetValidatorOutstandingRewards(ctx, del.GetValidatorAddr(), outstanding.Sub(rewards))
	feePool := k.GetFeePool(ctx)
	feePool.CommunityPool = feePool.CommunityPool.Add(remainder)
	k.SetFeePool(ctx, feePool)

	// add coins to user account
	if !coins.IsZero() {
		withdrawAddr := k.GetDelegatorWithdrawAddr(ctx, del.GetDelegatorAddr())
		if _, _, err := k.bankKeeper.AddCoins(ctx, withdrawAddr, coins); err != nil {
			return err
		}
	}

	// remove delegator starting info
	k.DeleteDelegatorStartingInfo(ctx, del.GetValidatorAddr(), del.GetDelegatorAddr())

	return nil
}
