package keeper

import (
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
	stake := delegation.GetShares().MulTruncate(validator.GetDelegatorShareExRate())
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
	difference := ending.CumulativeRewardRatio.Minus(starting.CumulativeRewardRatio)
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
	startingPeriod := startingInfo.PreviousPeriod
	stake := startingInfo.Stake

	// iterate through slashes and withdraw with calculated staking for sub-intervals
	// these offsets are dependent on *when* slashes happen - namely, in BeginBlock, after rewards are allocated...
	// ... so we don't reduce stake for slashes which happened in the *first* block, because the delegation wouldn't have existed
	startingHeight := startingInfo.Height + 1
	// ... or slashes which happened in *this* block, since they would have happened after reward allocation
	endingHeight := uint64(ctx.BlockHeight()) - 1
	if endingHeight >= startingHeight {
		k.IterateValidatorSlashEventsBetween(ctx, del.GetValidatorAddr(), startingHeight, endingHeight,
			func(height uint64, event types.ValidatorSlashEvent) (stop bool) {
				endingPeriod := event.ValidatorPeriod
				rewards = rewards.Plus(k.calculateDelegationRewardsBetween(ctx, val, startingPeriod, endingPeriod, stake))
				// note: necessary to truncate so we don't allow withdrawing more rewards than owed
				stake = stake.MulTruncate(sdk.OneDec().Sub(event.Fraction))
				startingPeriod = endingPeriod
				return false
			},
		)
	}

	// calculate rewards for final period
	rewards = rewards.Plus(k.calculateDelegationRewardsBetween(ctx, val, startingPeriod, endingPeriod, stake))

	return
}

func (k Keeper) withdrawDelegationRewards(ctx sdk.Context, val sdk.Validator, del sdk.Delegation) sdk.Error {

	// check existence of delegator starting info
	if !k.HasDelegatorStartingInfo(ctx, del.GetValidatorAddr(), del.GetDelegatorAddr()) {
		return types.ErrNoDelegationDistInfo(k.codespace)
	}

	// end current period and calculate rewards
	endingPeriod := k.incrementValidatorPeriod(ctx, val)
	rewards := k.calculateDelegationRewards(ctx, val, del, endingPeriod)

	// decrement reference count of starting period
	startingInfo := k.GetDelegatorStartingInfo(ctx, del.GetValidatorAddr(), del.GetDelegatorAddr())
	startingPeriod := startingInfo.PreviousPeriod
	k.decrementReferenceCount(ctx, del.GetValidatorAddr(), startingPeriod)

	// truncate coins, return remainder to community pool
	coins, remainder := rewards.TruncateDecimal()
	outstanding := k.GetOutstandingRewards(ctx)

	k.SetOutstandingRewards(ctx, outstanding.Minus(rewards))
	feePool := k.GetFeePool(ctx)
	feePool.CommunityPool = feePool.CommunityPool.Plus(remainder)
	k.SetFeePool(ctx, feePool)

	// add coins to user account
	withdrawAddr := k.GetDelegatorWithdrawAddr(ctx, del.GetDelegatorAddr())
	if _, _, err := k.bankKeeper.AddCoins(ctx, withdrawAddr, coins); err != nil {
		return err
	}

	// remove delegator starting info
	k.DeleteDelegatorStartingInfo(ctx, del.GetValidatorAddr(), del.GetDelegatorAddr())

	return nil
}
