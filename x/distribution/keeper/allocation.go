package keeper

import (
	"context"

	abci "github.com/cometbft/cometbft/abci/types"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// AllocateTokens performs reward and fee distribution to all validators based
// on the F1 fee distribution and Nakamoto bonus specifications.
func (k Keeper) AllocateTokens(ctx context.Context, totalPreviousPower int64, bondedVotes []abci.VoteInfo) error {
	// fetch and clear the collected fees for distribution, since this is
	// called in BeginBlock, collected fees will be from the previous block
	// (and distributed to the previous proposer)
	feeCollector := k.authKeeper.GetModuleAccount(ctx, k.feeCollectorName)
	feesCollectedInt := k.bankKeeper.GetAllBalances(ctx, feeCollector.GetAddress())
	feesCollected := sdk.NewDecCoinsFromCoins(feesCollectedInt...)

	// transfer collected fees to the distribution module account
	err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, k.feeCollectorName, types.ModuleName, feesCollectedInt)
	if err != nil {
		return err
	}

	// temporary workaround to keep CanWithdrawInvariant happy
	// general discussions here: https://github.com/cosmos/cosmos-sdk/issues/2906#issuecomment-441867634
	feePool, err := k.FeePool.Get(ctx)
	if err != nil {
		return err
	}

	if totalPreviousPower == 0 {
		feePool.CommunityPool = feePool.CommunityPool.Add(feesCollected...)
		return k.FeePool.Set(ctx, feePool)
	}

	// Get community tax and Nakamoto bonus ratio η
	communityTax, err := k.GetCommunityTax(ctx)
	if err != nil {
		return err
	}

	nb, err := k.GetNakamotoBonus(ctx)
	if err != nil {
		return err
	}

	nakamotoCoefficient, err := k.GetNakamotoBonusCoefficient(ctx)
	if err != nil {
		return err
	}

	// Compute total validator rewards (after community tax)
	voteMultiplier := math.LegacyOneDec().Sub(communityTax)
	validatorTotalReward := feesCollected.MulDecTruncate(voteMultiplier)
	nbPerValidator := sdk.NewDecCoins()

	if nb.Enabled {
		// Split reward into Proportional (PR_i) and Nakamoto Bonus (NB_i)
		nakamotoBonus := validatorTotalReward.MulDecTruncate(nakamotoCoefficient)
		validatorTotalReward = validatorTotalReward.Sub(nakamotoBonus)

		// Compute per-validator fixed Nakamoto bonus
		numValidators := int64(len(bondedVotes))
		if numValidators > 0 && !nakamotoBonus.IsZero() {
			// Distribute Nakamoto bonus across all denominations
			for _, coin := range nakamotoBonus {
				amount := coin.Amount.Quo(math.LegacyNewDec(numValidators))
				nbPerValidator = nbPerValidator.Add(sdk.NewDecCoinFromDec(coin.Denom, amount))
			}
		}
	}

	remaining := feesCollected

	// Distribute rewards to each validator
	for _, vote := range bondedVotes {
		validator, err := k.stakingKeeper.ValidatorByConsAddr(ctx, vote.Validator.Address)
		if err != nil {
			return err
		}

		// Compute proportional share based on voting power
		powerFraction := math.LegacyNewDec(vote.Validator.Power).Quo(math.LegacyNewDec(totalPreviousPower))
		proportional := validatorTotalReward.MulDecTruncate(powerFraction)

		// Add fixed Nakamoto bonus to proportional share
		reward := proportional.Add(nbPerValidator...)

		if err = k.AllocateTokensToValidator(ctx, validator, reward); err != nil {
			return err
		}

		remaining = remaining.Sub(reward)
	}

	// allocate community funding
	feePool.CommunityPool = feePool.CommunityPool.Add(remaining...)
	return k.FeePool.Set(ctx, feePool)
}

// AllocateTokensToValidator allocate tokens to a particular validator,
// splitting according to commission.
func (k Keeper) AllocateTokensToValidator(ctx context.Context, val stakingtypes.ValidatorI, tokens sdk.DecCoins) error {
	// split tokens between validator and delegators according to commission
	commission := tokens.MulDec(val.GetCommission())
	shared := tokens.Sub(commission)

	valBz, err := k.stakingKeeper.ValidatorAddressCodec().StringToBytes(val.GetOperator())
	if err != nil {
		return err
	}

	// update current commission
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCommission,
			sdk.NewAttribute(sdk.AttributeKeyAmount, commission.String()),
			sdk.NewAttribute(types.AttributeKeyValidator, val.GetOperator()),
		),
	)
	currentCommission, err := k.GetValidatorAccumulatedCommission(ctx, valBz)
	if err != nil {
		return err
	}

	currentCommission.Commission = currentCommission.Commission.Add(commission...)
	err = k.SetValidatorAccumulatedCommission(ctx, valBz, currentCommission)
	if err != nil {
		return err
	}

	// update current rewards
	currentRewards, err := k.GetValidatorCurrentRewards(ctx, valBz)
	if err != nil {
		return err
	}

	currentRewards.Rewards = currentRewards.Rewards.Add(shared...)
	err = k.SetValidatorCurrentRewards(ctx, valBz, currentRewards)
	if err != nil {
		return err
	}

	// update outstanding rewards
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRewards,
			sdk.NewAttribute(sdk.AttributeKeyAmount, tokens.String()),
			sdk.NewAttribute(types.AttributeKeyValidator, val.GetOperator()),
		),
	)

	outstanding, err := k.GetValidatorOutstandingRewards(ctx, valBz)
	if err != nil {
		return err
	}

	outstanding.Rewards = outstanding.Rewards.Add(tokens...)
	return k.SetValidatorOutstandingRewards(ctx, valBz, outstanding)
}
