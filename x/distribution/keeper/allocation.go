package keeper

import (
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// allocate fees handles distribution of the collected fees
func (k Keeper) AllocateTokens(ctx sdk.Context, sumPrecommitPower, totalPower int64, proposer sdk.ConsAddress, votes []abci.VoteInfo) {

	// fetch collected fees & fee pool
	feesCollectedInt := k.feeCollectionKeeper.GetCollectedFees(ctx)
	feesCollected := sdk.NewDecCoins(feesCollectedInt)
	feePool := k.GetFeePool(ctx)

	// clear collected fees, which will now be distributed
	k.feeCollectionKeeper.ClearCollectedFees(ctx)

	// temporary workaround to keep CanWithdrawInvariant happy
	// general discussions here: https://github.com/cosmos/cosmos-sdk/issues/2906#issuecomment-441867634
	if totalPower == 0 {
		feePool.CommunityPool = feePool.CommunityPool.Plus(feesCollected)
		k.SetFeePool(ctx, feePool)
		return
	}

	// calculate fraction votes
	fractionVotes := sdk.NewDec(sumPrecommitPower).Quo(sdk.NewDec(totalPower))

	// calculate proposer reward
	baseProposerReward := k.GetBaseProposerReward(ctx)
	bonusProposerReward := k.GetBonusProposerReward(ctx)
	proposerMultiplier := baseProposerReward.Add(bonusProposerReward.Mul(fractionVotes))
	proposerReward := feesCollected.MulDec(proposerMultiplier)

	// pay proposer
	proposerValidator := k.stakingKeeper.ValidatorByConsAddr(ctx, proposer)
	k.AllocateTokensToValidator(ctx, proposerValidator, proposerReward)
	remaining := feesCollected.Minus(proposerReward)

	// calculate fraction allocated to validators
	communityTax := k.GetCommunityTax(ctx)
	voteMultiplier := sdk.OneDec().Sub(proposerMultiplier).Sub(communityTax)

	// allocate tokens proportionally to voting power
	// TODO consider parallelizing later, ref https://github.com/cosmos/cosmos-sdk/pull/3099#discussion_r246276376
	for _, vote := range votes {
		validator := k.stakingKeeper.ValidatorByConsAddr(ctx, vote.Validator.Address)

		// TODO likely we should only reward validators who actually signed the block.
		// ref https://github.com/cosmos/cosmos-sdk/issues/2525#issuecomment-430838701
		powerFraction := sdk.NewDec(vote.Validator.Power).Quo(sdk.NewDec(totalPower))
		reward := feesCollected.MulDec(voteMultiplier).MulDec(powerFraction)
		k.AllocateTokensToValidator(ctx, validator, reward)
		remaining = remaining.Minus(reward)
	}

	// allocate community funding
	feePool.CommunityPool = feePool.CommunityPool.Plus(remaining)
	k.SetFeePool(ctx, feePool)

	// update outstanding rewards
	outstanding := k.GetOutstandingRewards(ctx)
	outstanding = outstanding.Plus(feesCollected.Minus(remaining))
	k.SetOutstandingRewards(ctx, outstanding)

}

// allocate tokens to a particular validator, splitting according to commission
func (k Keeper) AllocateTokensToValidator(ctx sdk.Context, val sdk.Validator, tokens sdk.DecCoins) {
	// split tokens between validator and delegators according to commission
	commission := tokens.MulDec(val.GetCommission())
	shared := tokens.Minus(commission)

	// update current commission
	currentCommission := k.GetValidatorAccumulatedCommission(ctx, val.GetOperator())
	currentCommission = currentCommission.Plus(commission)
	k.SetValidatorAccumulatedCommission(ctx, val.GetOperator(), currentCommission)

	// update current rewards
	currentRewards := k.GetValidatorCurrentRewards(ctx, val.GetOperator())
	currentRewards.Rewards = currentRewards.Rewards.Plus(shared)
	k.SetValidatorCurrentRewards(ctx, val.GetOperator(), currentRewards)
}
