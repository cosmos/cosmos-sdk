package keeper

import (
	"fmt"
	"sort"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (k Keeper) unionStrSlices(a, b []string) []string {
	m := make(map[string]bool)
	sort.Strings(a)
	sort.Strings(b)

	for _, item := range a {
		m[item] = true
	}

	for _, item := range b {
		if _, ok := m[item]; !ok {
			a = append(a, item)
		}
	}
	return a
}

// iterate the blacklisted delegators to gather a list of validators they're delegated to
func (k Keeper) GetTaintedValidators(ctx sdk.Context) []string {
	// get the list of blacklisted delegators
	blacklistedDelAddrs := k.GetParams(ctx).NoRewardsDelegatorAddresses
	k.Logger(ctx).Info("Blacklisted delegators", "addrs", blacklistedDelAddrs)
	// get the list of validators they're delegated to
	taintedVals := []string{}
	for _, delAddr := range blacklistedDelAddrs {

		// get delegator to be used to get its validators
		del, error := sdk.AccAddressFromBech32(delAddr)
		if error != nil {
			// TODO: panic?
			panic(error)
		}

		k.Logger(ctx).Info(fmt.Sprintf("...grabbing delegations by blacklisted del... %s", del.String()))
		// TODO replace with something like stakingKeeper.GetDelegatorVealidators()
		validators := []string{
			"stridevaloper1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrgpwsqm",
			"stridevaloper17kht2x2ped6qytr2kklevtvmxpw7wq9rcfud5c",
			"stridevaloper1nnurja9zt97huqvsfuartetyjx63tc5zrj5x9f",
			"stridevaloper1py0fvhdtq4au3d9l88rec6vyda3e0wttx9x92w",
			"stridevaloper1c5jnf370kaxnv009yhc3jt27f549l5u3edn747"}

		taintedVals = k.unionStrSlices(taintedVals, validators)
		k.Logger(ctx).Info(fmt.Sprintf("...updated taintedVals %s", taintedVals))
	}
	k.Logger(ctx).Info(fmt.Sprintf("TaintedVals are %s", taintedVals))
	return taintedVals
}

// get a validator's total blacklisted delegation power
// 		returns (totalPower, blacklistedPower)
func (k Keeper) GetBlacklistedPower(ctx sdk.Context, valAddr string) (int64, int64) {

	blacklistedDelAddrs := k.GetParams(ctx).NoRewardsDelegatorAddresses
	k.Logger(ctx).Info("Blacklisted delegators", "addrs", blacklistedDelAddrs)
	// get validator
	val, error := sdk.ValAddressFromBech32(valAddr)
	if error != nil {
		// TODO: panic?
		panic(error)
	}
	valObj := k.stakingKeeper.Validator(ctx, val)
	valTotPower := sdk.TokensToConsensusPower(valObj.GetTokens(), sdk.DefaultPowerReduction)

	valBlacklistedPower := int64(0)
	for _, delAddr := range blacklistedDelAddrs {
		// convert delAddrs to dels
		del, err := sdk.AccAddressFromBech32(delAddr)
		if err != nil {
			// TODO: panic?
			panic(err)
		}

		// add the delegation share to total
		delegation := k.stakingKeeper.Delegation(ctx, del, val)
		if delegation != nil {
			// TODO: why does TokensFromShares return a dec, when all tokens are ints? I truncate manually here -- is that safe?
			shares := delegation.GetShares()
			tokens := valObj.TokensFromShares(shares).TruncateInt()
			consPower := sdk.TokensToConsensusPower(tokens, sdk.DefaultPowerReduction)
			k.Logger(ctx).Info(fmt.Sprintf("... addr %s, shares %s, tokens %s consPower %d defPowerReduction %s", delegation.GetDelegatorAddr(),
				shares.String(), tokens.String(),
				consPower, sdk.DefaultPowerReduction.String()))
			// valObj.TokensFromShares(shares).Add(total)
			valBlacklistedPower = valBlacklistedPower + consPower
		}
	}
	k.Logger(ctx).Info(fmt.Sprintf("Total valBlacklistedPower is %d", valBlacklistedPower))
	return valTotPower, valBlacklistedPower
}

// helper
func (k Keeper) StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// AllocateTokens handles distribution of the collected fees
// bondedVotes is a list of (validator address, validator voted on last block flag) for all
// validators in the bonded set.
func (k Keeper) AllocateTokens(
	ctx sdk.Context, sumPreviousPrecommitPower, totalPreviousPower int64,
	previousProposer sdk.ConsAddress, bondedVotes []abci.VoteInfo,
) {

	logger := k.Logger(ctx)

	// get the blacklisted validators from the param store
	taintedVals := k.GetTaintedValidators(ctx)
	k.Logger(ctx).Info(fmt.Sprintf("Tainted validators are: %v", taintedVals))
	// deduct the power of the blacklisted validator from the total power (so that the others are upscaled proportionally!)
	valsBlacklistedPower := int64(0)
	taintedValBlacklistAmts := map[string]int64{}
	// TODO get the list of vals to iterate over from the blacklisted *delegators* so we don't iter all the vals
	for _, valAddr := range taintedVals {
		blacklistedValAddr, error := sdk.ValAddressFromBech32(valAddr)
		if error != nil {
			panic(error)
		}
		valTotalPower, valBlacklistedPower := k.GetBlacklistedPower(ctx, valAddr)
		valsBlacklistedPower += valBlacklistedPower
		taintedValBlacklistAmts[valAddr] = valBlacklistedPower
		k.Logger(ctx).Info(fmt.Sprintf("...tainted val %s has blacklistedpower: %d / %d", blacklistedValAddr, valBlacklistedPower, valTotalPower))
	}
	k.Logger(ctx).Info(fmt.Sprintf("Total valsBlacklistedPower is %d", valsBlacklistedPower))
	k.Logger(ctx).Info(fmt.Sprintf("TaintedValBlacklistAmts are %#v", taintedValBlacklistAmts))

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

	// allocate tokens proportionally to voting power
	// TODO consider parallelizing later, ref https://github.com/cosmos/cosmos-sdk/pull/3099#discussion_r246276376
	adjustedTotalPower := totalPreviousPower - valsBlacklistedPower
	for _, vote := range bondedVotes {

		validator := k.stakingKeeper.ValidatorByConsAddr(ctx, vote.Validator.Address)
		valAddr := validator.GetOperator().String()

		// validator's power begins at full power
		validatorPowerAdj := vote.Validator.Power

		// reduce the validator's power if they are tainted
		if k.StringInSlice(valAddr, taintedVals) {
			k.Logger(ctx).Info(fmt.Sprintf("...reducing val %s power: %d - %d ===> %d ", valAddr, validatorPowerAdj, taintedValBlacklistAmts[valAddr], validatorPowerAdj-taintedValBlacklistAmts[valAddr]))
			validatorPowerAdj -= taintedValBlacklistAmts[valAddr]
		}
		// TODO consider microslashing for missing votes.
		// ref https://github.com/cosmos/cosmos-sdk/issues/2525#issuecomment-430838701
		powerFraction := sdk.NewDec(validatorPowerAdj).QuoTruncate(sdk.NewDec(adjustedTotalPower))
		reward := feesCollected.MulDecTruncate(voteMultiplier).MulDecTruncate(powerFraction)

		k.AllocateTokensToValidator(ctx, validator, reward)
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
	currentRewards.Rewards = currentRewards.Rewards.Add(shared...)
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
	outstanding.Rewards = outstanding.Rewards.Add(tokens...)
	k.SetValidatorOutstandingRewards(ctx, val.GetOperator(), outstanding)
}
