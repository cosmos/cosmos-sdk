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
	proposerReward := k.GetProposerReward(ctx)
	proposerMultiplier := proposerReward.Mul(fractionVotes)
	proposerRewardAmount := feesCollected.MulDec(proposerMultiplier)

	// pay proposer
	proposerValidator := k.stakingKeeper.ValidatorByConsAddr(ctx, proposer)
	k.AllocateTokensToValidator(ctx, proposerValidator, proposerRewardAmount)
	remaining := feesCollected.Minus(proposerRewardAmount)

	// calculate fraction allocated to validators
	communityTax := k.GetCommunityTax(ctx)
	signerReward := k.GetSignerReward(ctx)
	signerMultiplier := signerReward
	bondedMultiplier := sdk.OneDec().Sub(proposerReward).Sub(communityTax).Sub(signerReward)

	// allocate tokens proportionally to voting power, each to signing / bonded
	for _, vote := range votes {
		var tokens sdk.DecCoins

		totalPowerFraction := sdk.NewDec(vote.Validator.Power).Quo(sdk.NewDec(totalPower))
		tokens = tokens.Plus(feesCollected.MulDec(bondedMultiplier).MulDec(totalPowerFraction))

		if vote.SignedLastBlock {
			signedPowerFraction := sdk.NewDec(vote.Validator.Power).Quo(sdk.NewDec(sumPrecommitPower))
			tokens = tokens.Plus(feesCollected.MulDec(signerMultiplier).MulDec(signedPowerFraction))
		}

		validator := k.stakingKeeper.ValidatorByConsAddr(ctx, vote.Validator.Address)
		k.AllocateTokensToValidator(ctx, validator, tokens)
		remaining = remaining.Minus(tokens)
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
