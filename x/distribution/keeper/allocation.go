package keeper

import (
	"fmt"
	"strconv"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// AllocateTokens handles distribution of the collected fees
// bondedVotes is a list of (validator address, validator voted on last block flag) for all
// validators in the bonded set.
func (k Keeper) AllocateTokens(
	ctx sdk.Context, sumPreviousPrecommitPower, totalPreviousPower int64,
	previousProposer sdk.ConsAddress, bondedVotes []abci.VoteInfo,
) {

	logger := k.Logger(ctx)

	// fetch values needed for blacklist logic
	// for block n, we are distributing rewards for block n-1
	// the validator set finalzied at n-3 signs n-1
	// see: https://github.com/tendermint/tendermint/pull/1815 and
	// https://github.com/tendermint/tendermint/blob/7b40167f58789803610747a4c385c0deee030f90/UPGRADING.md#validator-set-updates
	// for more details
	height := strconv.FormatInt(ctx.BlockHeight()-3, 10)
	fmt.Println("MOOSE")
	fmt.Println(height)
	blacklistedPower, found := k.GetBlacklistedPower(ctx, height)
	if !found {
		fmt.Println(blacklistedPower)
		k.Logger(ctx).Error(fmt.Sprintf("no blacklisted power found for current block height%s", height))
		return
	}
	totalBlacklistedPower := blacklistedPower.TotalBlacklistedPowerShare
	validatorBlacklistedPowers := blacklistedPower.ValidatorBlacklistedPowers
	if len(validatorBlacklistedPowers) == 0 {
		k.Logger(ctx).Error(fmt.Sprintf("no validator blacklisted power found for current block height%s", height))
		return
	}
	blacklistedPowerShareByValidator := k.GetBlacklistedPowerShareByValidator(ctx, validatorBlacklistedPowers)
	fmt.Println(totalBlacklistedPower.Quo(sdk.NewDec(totalPreviousPower)))
	totalWhitelistedPowerShare := sdk.NewDec(1).Sub(totalBlacklistedPower.Quo(sdk.NewDec(totalPreviousPower)))

	// fetch and clear the collected fees for distribution, since this is
	// called in BeginBlock, collected fees will be from the previous block
	// (and distributed to the previous proposer)
	feeCollector := k.authKeeper.GetModuleAccount(ctx, k.feeCollectorName)
	feesCollectedInt := k.bankKeeper.GetAllBalances(ctx, feeCollector.GetAddress())
	feesCollected := sdk.NewDecCoinsFromCoins(feesCollectedInt...)

	// transfer collected fees to the distribution module account
	err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, k.feeCollectorName, types.ModuleName, feesCollectedInt)
	if err != nil {
		panic(err)
	}

	// temporary workaround to keep CanWithdrawInvariant happy
	// general discussions here: https://github.com/cosmos/cosmos-sdk/issues/2906#issuecomment-441867634
	feePool := k.GetFeePool(ctx)
	if totalPreviousPower == 0 {
		feePool.CommunityPool = feePool.CommunityPool.Add(feesCollected...)
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

	// pay previous proposer
	remaining := feesCollected
	proposerValidator := k.stakingKeeper.ValidatorByConsAddr(ctx, previousProposer)

	if proposerValidator != nil {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeProposerReward,
				sdk.NewAttribute(sdk.AttributeKeyAmount, proposerReward.String()),
				sdk.NewAttribute(types.AttributeKeyValidator, proposerValidator.GetOperator().String()),
			),
		)

		k.AllocateTokensToValidator(ctx, proposerValidator, proposerReward)
		remaining = remaining.Sub(proposerReward)
	} else {
		// previous proposer can be unknown if say, the unbonding period is 1 block, so
		// e.g. a validator undelegates at block X, it's removed entirely by
		// block X+1's endblock, then X+2 we need to refer to the previous
		// proposer for X+1, but we've forgotten about them.
		logger.Error(fmt.Sprintf(
			"WARNING: Attempt to allocate proposer rewards to unknown proposer %s. "+
				"This should happen only if the proposer unbonded completely within a single block, "+
				"which generally should not happen except in exceptional circumstances (or fuzz testing). "+
				"We recommend you investigate immediately.",
			previousProposer.String()))
	}

	// calculate fraction allocated to validators
	communityTax := k.GetCommunityTax(ctx)
	voteMultiplier := sdk.OneDec().Sub(proposerMultiplier).Sub(communityTax)

	fmt.Println("voteMultiplier", voteMultiplier)
	// allocate tokens proportionally to voting power
	// TODO consider parallelizing later, ref https://github.com/cosmos/cosmos-sdk/pull/3099#discussion_r246276376
	adjustedTotalPower := sdk.NewDec(totalPreviousPower).Mul(totalWhitelistedPowerShare).RoundInt64() // TODO might rounding cause issues later?
	fmt.Println("adjustedTotalPower", adjustedTotalPower)
	fmt.Println("totalPreviousPower", totalPreviousPower)
	fmt.Println("totalWhitelistedPowerShare", totalWhitelistedPowerShare)
	// k.Logger(ctx).Info(fmt.Sprintf("\n... voteMultiplier %v, totalWhitelistedPowerShare %v, adjustedTotalPower %d \n", voteMultiplier, totalWhitelistedPowerShare, adjustedTotalPower))
	// fmt.Println("", adjustedTotalPower, voteMultiplier, totalWhitelistedPowerShare)
	for _, vote := range bondedVotes {
		validator := k.stakingKeeper.ValidatorByConsAddr(ctx, vote.Validator.Address)
		fmt.Println(validator.GetOperator())
		valAddr := validator.GetOperator().String()
		// k.Logger(ctx).Info(fmt.Sprintf("...%s", valAddr))

		var powerFraction sdk.Dec
		// reduce the validator's power if they are tainted
		if adjustedTotalPower != 0 { // If all we have is blacklisted delegations, process normally | TODO clean up this case
			valWhitelistedPowerShare := sdk.NewDec(1).Sub(blacklistedPowerShareByValidator[valAddr])
			validatorPowerAdj := sdk.NewDec(vote.Validator.Power).Mul(valWhitelistedPowerShare).RoundInt64()
			fmt.Println("vote.Validator.Power", vote.Validator.Power)
			fmt.Println("valWhitelistedPowerShare", valWhitelistedPowerShare)
			fmt.Println("validatorPowerAdj", validatorPowerAdj)
			fmt.Println("adjustedTotalPower", adjustedTotalPower)
			// k.Logger(ctx).Info(fmt.Sprintf("\t\t...tainted %s power: %d * %d ===> %d ", valAddr, vote.Validator.Power, valWhitelistedPowerShare, validatorPowerAdj))
			powerFraction = sdk.NewDec(validatorPowerAdj).QuoTruncate(sdk.NewDec(adjustedTotalPower))
		} else {
			// if not tainted use the untainted power fraction
			powerFraction = sdk.NewDec(vote.Validator.Power).QuoTruncate(sdk.NewDec(adjustedTotalPower))
		}
		// k.Logger(ctx).Info(fmt.Sprintf("\t\t...powerFraction %d", powerFraction))
		// TODO consider microslashing for missing votes.
		// ref https://github.com/cosmos/cosmos-sdk/issues/2525#issuecomment-430838701
		// if powerFraction.GT(sdk.OneDec()) {
		// 	powerFraction = powerFraction.Quo(sdk.NewDec(2))
		// }
		reward := feesCollected.MulDecTruncate(voteMultiplier).MulDecTruncate(powerFraction)

		k.AllocateTokensToValidator(ctx, validator, reward)
		// k.Logger(ctx).Info(fmt.Sprintf("\t\t... %#v to %s, remaining=%v", reward.AmountOf("ustrd"), validator.GetOperator().String(), remaining.AmountOf("ustrd")))
		remaining = remaining.Sub(reward)
	}

	// allocate community funding
	feePool.CommunityPool = feePool.CommunityPool.Add(remaining...)
	k.SetFeePool(ctx, feePool)
}

// AllocateTokensToValidator allocate tokens to a particular validator, splitting according to commission
func (k Keeper) AllocateTokensToValidator(ctx sdk.Context, val stakingtypes.ValidatorI, tokens sdk.DecCoins) {
	// split tokens between validator and delegators according to commission
	commission := tokens.MulDec(val.GetCommission())
	shared := tokens.Sub(commission)
	//log
	// k.Logger(ctx).Info(fmt.Sprintf("...2allocateTokensToValidator: val %s, amount %#v", val.GetOperator().String(), shared))

	// update current commission
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCommission,
			sdk.NewAttribute(sdk.AttributeKeyAmount, commission.String()),
			sdk.NewAttribute(types.AttributeKeyValidator, val.GetOperator().String()),
		),
	)
	currentCommission := k.GetValidatorAccumulatedCommission(ctx, val.GetOperator())
	currentCommission.Commission = currentCommission.Commission.Add(commission...)
	k.SetValidatorAccumulatedCommission(ctx, val.GetOperator(), currentCommission)

	// update current rewards
	currentRewards := k.GetValidatorCurrentRewards(ctx, val.GetOperator())
	// k.Logger(ctx).Info(fmt.Sprintf("...3allocateTokensToValidator: currentRewards %s, amount %#v", val.GetOperator().String(), currentRewards.Rewards))
	currentRewards.Rewards = currentRewards.Rewards.Add(shared...)
	// k.Logger(ctx).Info(fmt.Sprintf("...4allocateTokensToValidator: newRewards %s, amount %#v", val.GetOperator().String(), currentRewards.Rewards))
	k.SetValidatorCurrentRewards(ctx, val.GetOperator(), currentRewards)

	// update outstanding rewards
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRewards,
			sdk.NewAttribute(sdk.AttributeKeyAmount, tokens.String()),
			sdk.NewAttribute(types.AttributeKeyValidator, val.GetOperator().String()),
		),
	)
	outstanding := k.GetValidatorOutstandingRewards(ctx, val.GetOperator())
	// k.Logger(ctx).Info(fmt.Sprintf("    ...pre-outstanding %s, amount %d, adding %d tokens", val.GetOperator().String(), outstanding.Rewards.AmountOf("ustrd"), tokens.AmountOf("ustrd")))
	outstanding.Rewards = outstanding.Rewards.Add(tokens...)
	// k.Logger(ctx).Info(fmt.Sprintf("    ...post-outstanding: %s, amount %d", val.GetOperator().String(), outstanding.Rewards.AmountOf("ustrd")))
	k.SetValidatorOutstandingRewards(ctx, val.GetOperator(), outstanding)
}
