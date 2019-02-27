package keeper

import (
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO this is a hack
func (k Keeper) CalcWithdrawable(ctx sdk.Context, val sdk.Validator) sdk.DecCoins {
	ctx, _ = ctx.CacheContext()
	outstanding := k.GetValidatorOutstandingRewards(ctx, val.GetOperator())
	_ = k.WithdrawValidatorCommission(ctx, val.GetOperator())
	dels := k.stakingKeeper.GetAllSDKDelegations(ctx)
	for _, delegation := range dels {
		if delegation.GetValidatorAddr().String() == val.GetOperator().String() {
			//fmt.Printf("withdraw for delegator: %s\n", delegation.GetDelegatorAddr())
			_ = k.WithdrawDelegationRewards(ctx, delegation.GetDelegatorAddr(), delegation.GetValidatorAddr())
		}
	}
	remaining := k.GetValidatorOutstandingRewards(ctx, val.GetOperator())
	return outstanding.Sub(remaining)
}

// allocate fees handles distribution of the collected fees
func (k Keeper) AllocateTokens(ctx sdk.Context, sumPreviousPrecommitPower, totalPreviousPower int64,
	previousProposer sdk.ConsAddress, previousVotes []abci.VoteInfo) {

	// fetch and clear the collected fees for distribution, since this is
	// called in BeginBlock, collected fees will be from the previous block
	// (and distributed to the previous proposer)
	feesCollectedInt := k.feeCollectionKeeper.GetCollectedFees(ctx)
	feesCollected := sdk.NewDecCoins(feesCollectedInt)
	k.feeCollectionKeeper.ClearCollectedFees(ctx)

	// temporary workaround to keep CanWithdrawInvariant happy
	// general discussions here: https://github.com/cosmos/cosmos-sdk/issues/2906#issuecomment-441867634
	feePool := k.GetFeePool(ctx)
	if totalPreviousPower == 0 {
		feePool.CommunityPool = feePool.CommunityPool.Add(feesCollected)
		k.SetFeePool(ctx, feePool)
		return
	}

	// calculate fraction votes
	previousFractionVotes := sdk.NewDec(sumPreviousPrecommitPower).Quo(sdk.NewDec(totalPreviousPower))

	// calculate previous proposer reward
	baseProposerReward := k.GetBaseProposerReward(ctx)
	bonusProposerReward := k.GetBonusProposerReward(ctx)
	proposerMultiplier := baseProposerReward.Add(bonusProposerReward.MulTruncate(previousFractionVotes))
	proposerReward := feesCollected.MulDecTruncate(proposerMultiplier)
	fmt.Printf("\ndebug proposerReward: %v\n", proposerReward)

	// pay previous proposer
	remaining := feesCollected
	proposerValidator := k.stakingKeeper.ValidatorByConsAddr(ctx, previousProposer)
	fmt.Printf("debug proposerValidator: %v\n", proposerValidator)

	if proposerValidator != nil {
		k.AllocateTokensToValidator(ctx, proposerValidator, proposerReward)
		remaining = remaining.Sub(proposerReward)
	} else {
		// previous proposer can be unknown if say, the unbonding period is 1 block, so
		// e.g. a validator undelegates at block X, it's removed entirely by
		// block X+1's endblock, then X+2 we need to refer to the previous
		// proposer for X+1, but we've forgotten about them.
	}

	// calculate fraction allocated to validators
	communityTax := k.GetCommunityTax(ctx)
	voteMultiplier := sdk.OneDec().Sub(proposerMultiplier).Sub(communityTax)
	fmt.Printf("debug communityTax: %v\n", communityTax)
	fmt.Printf("debug proposerMultiplier: %v\n", proposerMultiplier)
	fmt.Printf("debug voteMultiplier: %v\n", voteMultiplier)

	// allocate tokens proportionally to voting power
	// TODO consider parallelizing later, ref https://github.com/cosmos/cosmos-sdk/pull/3099#discussion_r246276376
	for _, vote := range previousVotes {
		validator := k.stakingKeeper.ValidatorByConsAddr(ctx, vote.Validator.Address)

		// TODO likely we should only reward validators who actually signed the block.
		// ref https://github.com/cosmos/cosmos-sdk/issues/2525#issuecomment-430838701
		powerFraction := sdk.NewDec(vote.Validator.Power).QuoTruncate(sdk.NewDec(totalPreviousPower))
		reward := feesCollected.MulDecTruncate(voteMultiplier).MulDecTruncate(powerFraction)
		k.AllocateTokensToValidator(ctx, validator, reward)
		remaining = remaining.Sub(reward)
	}

	// allocate community funding
	feePool.CommunityPool = feePool.CommunityPool.Add(remaining)
	k.SetFeePool(ctx, feePool)

}

// allocate tokens to a particular validator, splitting according to commission
func (k Keeper) AllocateTokensToValidator(ctx sdk.Context, val sdk.Validator, tokens sdk.DecCoins) {

	withdrawablePrior := k.CalcWithdrawable(ctx, val)
	//fmt.Printf("allocating %v tokens to validator %s, prior withdrawable: %v\n",
	//tokens, val.GetOperator(), withdrawablePrior)

	//if val.GetOperator().String() == "cosmosvaloper1qu379fd7lzvl9pfwclw2984n99dfqdgxjypypy" {
	//fmt.Printf("allocate to validator: val %+v, tokens %v\n", val, tokens)
	//}

	// split tokens between validator and delegators according to commission
	commission := tokens.MulDec(val.GetCommission())
	shared := tokens.Sub(commission)

	// update current commission
	currentCommission := k.GetValidatorAccumulatedCommission(ctx, val.GetOperator())
	currentCommission = currentCommission.Add(commission)
	k.SetValidatorAccumulatedCommission(ctx, val.GetOperator(), currentCommission)

	// update current rewards
	currentRewards := k.GetValidatorCurrentRewards(ctx, val.GetOperator())
	currentRewards.Rewards = currentRewards.Rewards.Add(shared)
	k.SetValidatorCurrentRewards(ctx, val.GetOperator(), currentRewards)

	// update outstanding rewards
	outstanding := k.GetValidatorOutstandingRewards(ctx, val.GetOperator())
	outstanding = outstanding.Add(tokens)
	k.SetValidatorOutstandingRewards(ctx, val.GetOperator(), outstanding)

	withdrawablePost := k.CalcWithdrawable(ctx, val)
	if withdrawablePost.Sub(withdrawablePrior)[0].IsGT(tokens[0]) {
		msg := fmt.Sprintf("greater withdraw allowed than allocated: validator %s, allowed: %v, allocated %v\n",
			val.GetOperator(), withdrawablePost.Sub(withdrawablePrior), tokens)
		panic(msg)
	}
}
