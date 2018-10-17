package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	types "github.com/cosmos/cosmos-sdk/x/stake/types"
)

// Slash a validator for an infraction committed at a known height
// Find the contributing stake at that height and burn the specified slashFactor
// of it, updating unbonding delegation & redelegations appropriately
//
// CONTRACT:
//    slashFactor is non-negative
// CONTRACT:
//    Infraction committed equal to or less than an unbonding period in the past,
//    so all unbonding delegations and redelegations from that height are stored
// CONTRACT:
//    Slash will not slash unbonded validators (for the above reason)
// CONTRACT:
//    Infraction committed at the current height or at a past height,
//    not at a height in the future
func (k Keeper) Slash(ctx sdk.Context, consAddr sdk.ConsAddress, infractionHeight int64, power int64, slashFactor sdk.Dec) {
	logger := ctx.Logger().With("module", "x/stake")

	if slashFactor.LT(sdk.ZeroDec()) {
		panic(fmt.Errorf("attempted to slash with a negative slash factor: %v", slashFactor))
	}

	// Amount of slashing = slash slashFactor * power at time of infraction
	slashAmount := sdk.NewDec(power).Mul(slashFactor)
	// ref https://github.com/cosmos/cosmos-sdk/issues/1348
	// ref https://github.com/cosmos/cosmos-sdk/issues/1471

	validator, found := k.GetValidatorByConsAddr(ctx, consAddr)
	if !found {
		// If not found, the validator must have been overslashed and removed - so we don't need to do anything
		// NOTE:  Correctness dependent on invariant that unbonding delegations / redelegations must also have been completely
		//        slashed in this case - which we don't explicitly check, but should be true.
		// Log the slash attempt for future reference (maybe we should tag it too)
		logger.Error(fmt.Sprintf(
			"WARNING: Ignored attempt to slash a nonexistent validator with address %s, we recommend you investigate immediately",
			consAddr))
		return
	}

	// should not be slashing unbonded
	if validator.Status == sdk.Unbonded {
		panic(fmt.Sprintf("should not be slashing unbonded validator: %s", validator.GetOperator()))
	}

	operatorAddress := validator.GetOperator()

	// Track remaining slash amount for the validator
	// This will decrease when we slash unbondings and
	// redelegations, as that stake has since unbonded
	remainingSlashAmount := slashAmount

	switch {
	case infractionHeight > ctx.BlockHeight():

		// Can't slash infractions in the future
		panic(fmt.Sprintf(
			"impossible attempt to slash future infraction at height %d but we are at height %d",
			infractionHeight, ctx.BlockHeight()))

	case infractionHeight == ctx.BlockHeight():

		// Special-case slash at current height for efficiency - we don't need to look through unbonding delegations or redelegations
		logger.Info(fmt.Sprintf(
			"slashing at current height %d, not scanning unbonding delegations & redelegations",
			infractionHeight))

	case infractionHeight < ctx.BlockHeight():

		// Iterate through unbonding delegations from slashed validator
		unbondingDelegations := k.GetUnbondingDelegationsFromValidator(ctx, operatorAddress)
		for _, unbondingDelegation := range unbondingDelegations {
			amountSlashed := k.slashUnbondingDelegation(ctx, unbondingDelegation, infractionHeight, slashFactor)
			if amountSlashed.IsZero() {
				continue
			}
			remainingSlashAmount = remainingSlashAmount.Sub(amountSlashed)
		}

		// Iterate through redelegations from slashed validator
		redelegations := k.GetRedelegationsFromValidator(ctx, operatorAddress)
		for _, redelegation := range redelegations {
			amountSlashed := k.slashRedelegation(ctx, validator, redelegation, infractionHeight, slashFactor)
			if amountSlashed.IsZero() {
				continue
			}
			remainingSlashAmount = remainingSlashAmount.Sub(amountSlashed)
		}
	}

	// cannot decrease balance below zero
	tokensToBurn := sdk.MinDec(remainingSlashAmount, validator.Tokens)

	// burn validator's tokens and update the validator
	validator = k.RemoveValidatorTokens(ctx, validator, tokensToBurn)
	pool := k.GetPool(ctx)
	pool.LooseTokens = pool.LooseTokens.Sub(tokensToBurn)
	k.SetPool(ctx, pool)

	// remove validator if it has no more tokens
	if validator.Tokens.IsZero() && validator.Status != sdk.Bonded {
		// if bonded, we must remove in ApplyAndReturnValidatorSetUpdates instead
		k.RemoveValidator(ctx, validator.OperatorAddr)
	}

	// Log that a slash occurred!
	logger.Info(fmt.Sprintf(
		"validator %s slashed by slash factor of %s; burned %v tokens",
		validator.GetOperator(), slashFactor.String(), tokensToBurn))

	// TODO Return event(s), blocked on https://github.com/tendermint/tendermint/pull/1803
	return
}

// jail a validator
func (k Keeper) Jail(ctx sdk.Context, consAddr sdk.ConsAddress) {
	validator := k.mustGetValidatorByConsAddr(ctx, consAddr)
	k.jailValidator(ctx, validator)
	logger := ctx.Logger().With("module", "x/stake")
	logger.Info(fmt.Sprintf("validator %s jailed", consAddr))
	// TODO Return event(s), blocked on https://github.com/tendermint/tendermint/pull/1803
	return
}

// unjail a validator
func (k Keeper) Unjail(ctx sdk.Context, consAddr sdk.ConsAddress) {
	validator := k.mustGetValidatorByConsAddr(ctx, consAddr)
	k.unjailValidator(ctx, validator)
	logger := ctx.Logger().With("module", "x/stake")
	logger.Info(fmt.Sprintf("validator %s unjailed", consAddr))
	// TODO Return event(s), blocked on https://github.com/tendermint/tendermint/pull/1803
	return
}

// slash an unbonding delegation and update the pool
// return the amount that would have been slashed assuming
// the unbonding delegation had enough stake to slash
// (the amount actually slashed may be less if there's
// insufficient stake remaining)
func (k Keeper) slashUnbondingDelegation(ctx sdk.Context, unbondingDelegation types.UnbondingDelegation,
	infractionHeight int64, slashFactor sdk.Dec) (slashAmount sdk.Dec) {

	now := ctx.BlockHeader().Time

	// If unbonding started before this height, stake didn't contribute to infraction
	if unbondingDelegation.CreationHeight < infractionHeight {
		return sdk.ZeroDec()
	}

	if unbondingDelegation.MinTime.Before(now) {
		// Unbonding delegation no longer eligible for slashing, skip it
		// TODO Settle and delete it automatically?
		return sdk.ZeroDec()
	}

	// Calculate slash amount proportional to stake contributing to infraction
	slashAmount = slashFactor.MulInt(unbondingDelegation.InitialBalance.Amount)

	// Don't slash more tokens than held
	// Possible since the unbonding delegation may already
	// have been slashed, and slash amounts are calculated
	// according to stake held at time of infraction
	unbondingSlashAmount := sdk.MinInt(slashAmount.RoundInt(), unbondingDelegation.Balance.Amount)

	// Update unbonding delegation if necessary
	if !unbondingSlashAmount.IsZero() {
		unbondingDelegation.Balance.Amount = unbondingDelegation.Balance.Amount.Sub(unbondingSlashAmount)
		k.SetUnbondingDelegation(ctx, unbondingDelegation)
		pool := k.GetPool(ctx)

		// Burn loose tokens
		// Ref https://github.com/cosmos/cosmos-sdk/pull/1278#discussion_r198657760
		pool.LooseTokens = pool.LooseTokens.Sub(slashAmount)
		k.SetPool(ctx, pool)
	}

	return
}

// slash a redelegation and update the pool
// return the amount that would have been slashed assuming
// the unbonding delegation had enough stake to slash
// (the amount actually slashed may be less if there's
// insufficient stake remaining)
// nolint: unparam
func (k Keeper) slashRedelegation(ctx sdk.Context, validator types.Validator, redelegation types.Redelegation,
	infractionHeight int64, slashFactor sdk.Dec) (slashAmount sdk.Dec) {

	now := ctx.BlockHeader().Time

	// If redelegation started before this height, stake didn't contribute to infraction
	if redelegation.CreationHeight < infractionHeight {
		return sdk.ZeroDec()
	}

	if redelegation.MinTime.Before(now) {
		// Redelegation no longer eligible for slashing, skip it
		// TODO Delete it automatically?
		return sdk.ZeroDec()
	}

	// Calculate slash amount proportional to stake contributing to infraction
	slashAmount = slashFactor.MulInt(redelegation.InitialBalance.Amount)

	// Don't slash more tokens than held
	// Possible since the redelegation may already
	// have been slashed, and slash amounts are calculated
	// according to stake held at time of infraction
	redelegationSlashAmount := sdk.MinInt(slashAmount.RoundInt(), redelegation.Balance.Amount)

	// Update redelegation if necessary
	if !redelegationSlashAmount.IsZero() {
		redelegation.Balance.Amount = redelegation.Balance.Amount.Sub(redelegationSlashAmount)
		k.SetRedelegation(ctx, redelegation)
	}

	// Unbond from target validator
	sharesToUnbond := slashFactor.Mul(redelegation.SharesDst)
	if !sharesToUnbond.IsZero() {
		delegation, found := k.GetDelegation(ctx, redelegation.DelegatorAddr, redelegation.ValidatorDstAddr)
		if !found {
			// If deleted, delegation has zero shares, and we can't unbond any more
			return slashAmount
		}
		if sharesToUnbond.GT(delegation.Shares) {
			sharesToUnbond = delegation.Shares
		}
		tokensToBurn, err := k.unbond(ctx, redelegation.DelegatorAddr, redelegation.ValidatorDstAddr, sharesToUnbond)
		if err != nil {
			panic(fmt.Errorf("error unbonding delegator: %v", err))
		}

		// Burn loose tokens
		pool := k.GetPool(ctx)
		pool.LooseTokens = pool.LooseTokens.Sub(tokensToBurn)
		k.SetPool(ctx, pool)
	}

	return slashAmount
}
