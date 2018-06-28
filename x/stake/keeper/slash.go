package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	types "github.com/cosmos/cosmos-sdk/x/stake/types"
	crypto "github.com/tendermint/go-crypto"
)

// slash an unbonding delegation
func (k Keeper) slashUnbondingDelegation(ctx sdk.Context, unbondingDelegation types.UnbondingDelegation, infractionHeight int64, fraction sdk.Rat) (slashAmount sdk.Rat) {
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
	slashAmount = sdk.NewRatFromInt(unbondingDelegation.InitialBalance.Amount, sdk.OneInt()).Mul(fraction)

	// Don't slash more tokens than held
	slashAmountInt := slashAmount.EvaluateInt()
	if slashAmountInt.GT(unbondingDelegation.Balance.Amount) {
		slashAmountInt = unbondingDelegation.Balance.Amount
	}

	// Update unbonding delegation
	unbondingDelegation.Balance.Amount = unbondingDelegation.Balance.Amount.Sub(slashAmountInt)
	k.SetUnbondingDelegation(ctx, unbondingDelegation)

	return
}

// slash a redelegation
func (k Keeper) slashRedelegation(ctx sdk.Context, redelegation types.Redelegation, infractionHeight int64, fraction sdk.Rat) (slashAmount sdk.Rat, tokensToBurn int64) {
	now := ctx.BlockHeader().Time

	// If redelegation started before this height, stake didn't contribute to infraction
	if redelegation.CreationHeight < infractionHeight {
		return sdk.ZeroRat(), 0
	}

	if redelegation.MinTime < now {
		// Redelegation no longer eligible for slashing, skip it
		// TODO Delete it automatically?
		return sdk.ZeroRat(), 0
	}

	// Calculate slash amount proportional to stake contributing to infraction
	slashAmount = sdk.NewRatFromInt(redelegation.InitialBalance.Amount, sdk.OneInt()).Mul(fraction)

	// Don't slash more tokens than held
	slashAmountInt := slashAmount.EvaluateInt()
	if slashAmountInt.GT(redelegation.Balance.Amount) {
		slashAmountInt = redelegation.Balance.Amount
	}

	// Update redelegation
	redelegation.Balance.Amount = redelegation.Balance.Amount.Sub(slashAmountInt)
	k.SetRedelegation(ctx, redelegation)

	// Unbond from target validator
	sharesToUnbond := fraction.Mul(redelegation.SharesDst)
	tokensToBurn, err := k.unbond(ctx, redelegation.DelegatorAddr, redelegation.ValidatorDstAddr, sharesToUnbond)
	if err != nil {
		panic(fmt.Errorf("error unbonding delegator: %v", err))
	}

	return slashAmount, tokensToBurn
}

// Slash a validator for an infraction committed at a known height
// Find the contributing stake at that height and burn the specified fraction
// of it, updating unbonding delegation & redelegations appropriately
//
// CONTRACT: Infraction committed equal to or less than an unbonding period in the past,
// so all unbonding delegations and redelegations from that height are stored
func (k Keeper) Slash(ctx sdk.Context, pubkey crypto.PubKey, infractionHeight int64, power int64, fraction sdk.Rat) {
	logger := ctx.Logger().With("module", "x/stake")

	// Amount of slashing = slash fraction * power at time of infraction
	slashAmount := sdk.NewRat(power).Mul(fraction)
	// hmm, https://github.com/cosmos/cosmos-sdk/issues/1348

	validator, found := k.GetValidatorByPubKey(ctx, pubkey)
	if !found {
		panic(fmt.Errorf("attempted to slash a nonexistent validator with address %s", pubkey.Address()))
	}
	ownerAddress := validator.GetOwner()

	// Track remaining slash amount
	remainingSlashAmount := slashAmount

	// Get the current pool so we can update it as we go
	pool := k.GetPool(ctx)

	switch {
	case infractionHeight > ctx.BlockHeight():
		// Can't slash infractions in the future
		panic(fmt.Sprintf("impossible attempt to slash future infraction at height %d but we are at height %d", infractionHeight, ctx.BlockHeight()))

	case infractionHeight == ctx.BlockHeight():
		// Special-case slash at current height for efficiency - we don't need to look through unbonding delegations or redelegations
		logger.Info(fmt.Sprintf("Slashing at current height %d, not scanning unbonding delegations & redelegations", infractionHeight))

	case infractionHeight < ctx.BlockHeight():
		// Iterate through unbonding delegations from slashed validator
		unbondingDelegations := k.GetUnbondingDelegationsFromValidator(ctx, ownerAddress)
		for _, unbondingDelegation := range unbondingDelegations {
			amountSlashed := k.slashUnbondingDelegation(ctx, unbondingDelegation, infractionHeight, fraction)
			remainingSlashAmount = remainingSlashAmount.Sub(amountSlashed)
			// Burn unbonding tokens
			// Ref https://github.com/cosmos/cosmos-sdk/pull/1278#discussion_r198657760
			pool.LooseTokens -= amountSlashed.EvaluateInt().Int64()
		}

		// Iterate through redelegations from slashed validator
		redelegations := k.GetRedelegationsFromValidator(ctx, ownerAddress)
		for _, redelegation := range redelegations {
			amountSlashed, tokensToBurn := k.slashRedelegation(ctx, redelegation, infractionHeight, fraction)
			remainingSlashAmount = remainingSlashAmount.Sub(amountSlashed)
			// Burn bonded tokens
			pool.BondedTokens -= tokensToBurn
		}

	}

	sharesToRemove := remainingSlashAmount
	// Cannot decrease balance below zero
	if sharesToRemove.GT(validator.PoolShares.Amount) {
		sharesToRemove = validator.PoolShares.Amount
	}

	validator, pool, burned := validator.RemovePoolShares(pool, sharesToRemove) // remove shares from the validator
	pool.LooseTokens -= burned                                                  // burn tokens
	k.SetPool(ctx, pool)                                                        // update the pool
	k.UpdateValidator(ctx, validator)                                           // update the validator, possibly kicking it out

	// Log that a slash occurred!
	logger.Info(fmt.Sprintf("Validator %s slashed by fraction %v, removed %v shares and burned %d tokens", pubkey.Address(), fraction, sharesToRemove, burned))

	// TODO Return event(s), blocked on https://github.com/tendermint/tendermint/pull/1803
	return
}

// revoke a validator
func (k Keeper) Revoke(ctx sdk.Context, pubkey crypto.PubKey) {
	k.setRevoked(ctx, pubkey, true)
	logger := ctx.Logger().With("module", "x/stake")
	logger.Info(fmt.Sprintf("Validator %s revoked", pubkey.Address()))
	return
}

// unrevoke a validator
func (k Keeper) Unrevoke(ctx sdk.Context, pubkey crypto.PubKey) {
	k.setRevoked(ctx, pubkey, false)
	logger := ctx.Logger().With("module", "x/stake")
	logger.Info(fmt.Sprintf("Validator %s unrevoked", pubkey.Address()))
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
