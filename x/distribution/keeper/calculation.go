package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

func (k Keeper) initializeValidator(ctx sdk.Context, val sdk.Validator) {
	// Set initial historical rewards (period 0)
	k.setValidatorHistoricalRewards(ctx, val.GetOperator(), 0, types.ValidatorHistoricalRewards{})

	// Set current rewards (starting at period 1)
	k.setValidatorCurrentRewards(ctx, val.GetOperator(), types.ValidatorCurrentRewards{
		Rewards: sdk.DecCoins{},
		Period:  1,
	})

	// Set accumulated commission
	k.setValidatorAccumulatedCommission(ctx, val.GetOperator(), types.ValidatorAccumulatedCommission{})
}

func (k Keeper) incrementValidatorPeriod(ctx sdk.Context, val sdk.Validator) uint64 {
	// Fetch current rewards
	rewards := k.GetValidatorCurrentRewards(ctx, val.GetOperator())

	// Calculate current ratio
	var current sdk.DecCoins
	if val.GetPower().IsZero() {
		// this can happen after redelegations are slashed
		current = sdk.DecCoins{}
		// TODO: Add to the community pool?
	} else {
		current = rewards.Rewards.QuoDec(sdk.NewDecFromInt(val.GetPower()))
	}

	// Fetch historical rewards for last period
	historical := k.GetValidatorHistoricalRewards(ctx, val.GetOperator(), rewards.Period-1)

	// Set new historical rewards
	k.setValidatorHistoricalRewards(ctx, val.GetOperator(), rewards.Period, historical.Plus(current))

	// Set current rewards, incrementing period by 1
	newPeriod := rewards.Period + 1
	k.setValidatorCurrentRewards(ctx, val.GetOperator(), types.ValidatorCurrentRewards{
		Rewards: sdk.DecCoins{},
		Period:  newPeriod,
	})

	return rewards.Period
}

func (k Keeper) initializeDelegation(ctx sdk.Context, val sdk.ValAddress, del sdk.AccAddress) {
	period := k.GetValidatorCurrentRewards(ctx, val).Period
	k.setDelegatorStartingInfo(ctx, val, del, types.DelegatorStartingInfo{
		PreviousPeriod: period,
		Stake:          sdk.ZeroInt(),
	})
}

func (k Keeper) withdrawDelegationRewards(ctx sdk.Context, val sdk.Validator, del sdk.Delegation) sdk.Error {
	endedPeriod := k.incrementValidatorPeriod(ctx, val)
	present := k.GetValidatorHistoricalRewards(ctx, del.GetValidatorAddr(), endedPeriod)
	startingInfo := k.GetDelegatorStartingInfo(ctx, del.GetValidatorAddr(), del.GetDelegatorAddr())
	historical := k.GetValidatorHistoricalRewards(ctx, del.GetValidatorAddr(), startingInfo.PreviousPeriod)
	difference := present.Minus(historical)
	rewards := difference.MulDec(sdk.NewDecFromInt(startingInfo.Stake))

	// Truncate coins, return remainder to community pool
	coins, remainder := rewards.TruncateDecimal()
	outstanding := k.GetOutstandingRewards(ctx)
	k.SetOutstandingRewards(ctx, outstanding.Minus(rewards))
	feePool := k.GetFeePool(ctx)
	feePool.CommunityPool = feePool.CommunityPool.Plus(remainder)
	k.SetFeePool(ctx, feePool)

	// Add coins to user account
	withdrawAddr := k.GetDelegatorWithdrawAddr(ctx, del.GetDelegatorAddr())
	if _, _, err := k.bankKeeper.AddCoins(ctx, withdrawAddr, coins); err != nil {
		return err
	}

	return nil
}
