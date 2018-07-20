package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	types "github.com/cosmos/cosmos-sdk/x/stake/types"
	"github.com/tendermint/tendermint/crypto"
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
//    Infraction committed at the current height or at a past height,
//    not at a height in the future
func (k Keeper) Slash(ctx sdk.Context, pubkey crypto.PubKey, infractionHeight int64, power int64, slashFactor sdk.Rat) {
	logger := ctx.Logger().With("module", "x/stake")

	if slashFactor.LT(sdk.ZeroRat()) {
		panic(fmt.Errorf("attempted to slash with a negative slashFactor: %v", slashFactor))
	}

	// Amount of slashing = slash slashFactor * power at time of infraction
	slashAmount := sdk.NewRat(power).Mul(slashFactor)
	// ref https://github.com/cosmos/cosmos-sdk/issues/1348
	// ref https://github.com/cosmos/cosmos-sdk/issues/1471

	validator, found := k.GetValidatorByPubKey(ctx, pubkey)
	if !found {
		// If not found, the validator must have been overslashed and removed - so we don't need to do anything
		// NOTE:  Correctness dependent on invariant that unbonding delegations / redelegations must also have been completely
		//        slashed in this case - which we don't explicitly check, but should be true.
		// Log the slash attempt for future reference (maybe we should tag it too)
		logger.Error(fmt.Sprintf(
			"WARNING: Ignored attempt to slash a nonexistent validator with address %s, we recommend you investigate immediately",
			pubkey.Address()))
		return
	}
	ownerAddress := validator.GetOwner()

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
			"Slashing at current height %d, not scanning unbonding delegations & redelegations",
			infractionHeight))

	case infractionHeight < ctx.BlockHeight():

		// Iterate through unbonding delegations from slashed validator
		unbondingDelegations := k.GetUnbondingDelegationsFromValidator(ctx, ownerAddress)
		for _, unbondingDelegation := range unbondingDelegations {
			amountSlashed := k.slashUnbondingDelegation(ctx, unbondingDelegation, infractionHeight, slashFactor)
			if amountSlashed.IsZero() {
				continue
			}
			remainingSlashAmount = remainingSlashAmount.Sub(amountSlashed)
		}

		// Iterate through redelegations from slashed validator
		redelegations := k.GetRedelegationsFromValidator(ctx, ownerAddress)
		for _, redelegation := range redelegations {
			amountSlashed := k.slashRedelegation(ctx, validator, redelegation, infractionHeight, slashFactor)
			if amountSlashed.IsZero() {
				continue
			}
			remainingSlashAmount = remainingSlashAmount.Sub(amountSlashed)
		}
	}

	// Cannot decrease balance below zero
	tokensToBurn := sdk.MinRat(remainingSlashAmount, validator.Tokens)

	// Get the current pool
	pool := k.GetPool(ctx)
	// remove tokens from the validator
	validator, pool = validator.RemoveTokens(pool, tokensToBurn)
	// burn tokens
	pool.LooseTokens = pool.LooseTokens.Sub(tokensToBurn)
	// update the pool
	k.SetPool(ctx, pool)
	// update the validator, possibly kicking it out
	validator = k.UpdateValidator(ctx, validator)
	// remove validator if it has been reduced to zero shares
	if validator.Tokens.IsZero() {
		k.RemoveValidator(ctx, validator.Owner)
	}

	// Log that a slash occurred!
	logger.Info(fmt.Sprintf(
		"Validator %s slashed by slashFactor %v, burned %v tokens",
		pubkey.Address(), slashFactor, tokensToBurn))

	// TODO Return event(s), blocked on https://github.com/tendermint/tendermint/pull/1803
	return
}

// revoke a validator
func (k Keeper) Revoke(ctx sdk.Context, pubkey crypto.PubKey) {
	k.setRevoked(ctx, pubkey, true)
	logger := ctx.Logger().With("module", "x/stake")
	logger.Info(fmt.Sprintf("Validator %s revoked", pubkey.Address()))
	// TODO Return event(s), blocked on https://github.com/tendermint/tendermint/pull/1803
	return
}

// unrevoke a validator
func (k Keeper) Unrevoke(ctx sdk.Context, pubkey crypto.PubKey) {
	k.setRevoked(ctx, pubkey, false)
	logger := ctx.Logger().With("module", "x/stake")
	logger.Info(fmt.Sprintf("Validator %s unrevoked", pubkey.Address()))
	// TODO Return event(s), blocked on https://github.com/tendermint/tendermint/pull/1803
	return
}

// set the revoked flag on a validator
func (k Keeper) setRevoked(ctx sdk.Context, pubkey crypto.PubKey, revoked bool) {
	validator, found := k.GetValidatorByPubKey(ctx, pubkey)
	if !found {
		panic(fmt.Errorf("Validator with pubkey %s not found, cannot set revoked to %v", pubkey, revoked))
	}
	validator.Revoked = revoked
	k.UpdateValidator(ctx, validator) // update validator, possibly unbonding or bonding it
	return
}

// slash an unbonding delegation and update the pool
// return the amount that would have been slashed assuming
// the unbonding delegation had enough stake to slash
// (the amount actually slashed may be less if there's
// insufficient stake remaining)
func (k Keeper) slashUnbondingDelegation(ctx sdk.Context, unbondingDelegation types.UnbondingDelegation,
	infractionHeight int64, slashFactor sdk.Rat) (slashAmount sdk.Rat) {

	now := ctx.BlockHeader().Time

	// If unbonding started before this height, stake didn't contribute to infraction
	if unbondingDelegation.CreationHeight < infractionHeight {
		return sdk.ZeroRat()
	}

	if unbondingDelegation.MinTime < now {
		// Unbonding delegation no longer eligible for slashing, skip it
		// TODO Settle and delete it automatically?
		return sdk.ZeroRat()
	}

	// Calculate slash amount proportional to stake contributing to infraction
	slashAmount = sdk.NewRatFromInt(unbondingDelegation.InitialBalance.Amount, sdk.OneInt()).Mul(slashFactor)

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
func (k Keeper) slashRedelegation(ctx sdk.Context, validator types.Validator, redelegation types.Redelegation,
	infractionHeight int64, slashFactor sdk.Rat) (slashAmount sdk.Rat) {

	now := ctx.BlockHeader().Time

	// If redelegation started before this height, stake didn't contribute to infraction
	if redelegation.CreationHeight < infractionHeight {
		return sdk.ZeroRat()
	}

	if redelegation.MinTime < now {
		// Redelegation no longer eligible for slashing, skip it
		// TODO Delete it automatically?
		return sdk.ZeroRat()
	}

	// Calculate slash amount proportional to stake contributing to infraction
	slashAmount = sdk.NewRatFromInt(redelegation.InitialBalance.Amount, sdk.OneInt()).Mul(slashFactor)

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
