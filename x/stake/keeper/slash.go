package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	crypto "github.com/tendermint/go-crypto"
)

// slash a validator
func (k Keeper) Slash(ctx sdk.Context, pubkey crypto.PubKey, height int64, power int64, fraction sdk.Rat) {

	// Amount of slashing = slash fraction * power at time of equivocation
	slashAmount := sdk.NewRat(power).Mul(fraction)
	// hmm, https://github.com/cosmos/cosmos-sdk/issues/1348

	// Current timestamp
	now := ctx.BlockHeader().Time

	validator, found := k.GetValidatorByPubKey(ctx, pubkey)
	if !found {
		panic(fmt.Errorf("Attempted to slash a nonexistent validator with address %s", pubkey.Address()))
	}
	address := pubkey.Address()

	// Track remaining slash amount
	remainingSlashAmount := slashAmount

	// Iterate through unbonding delegations from slashed validator
	unbondingDelegations := k.GetUnbondingDelegationsFromValidator(ctx, address)
	for _, unbondingDelegation := range unbondingDelegations {
		if unbondingDelegation.MinTime < now {
			// TODO Delete element?
			continue
		}

		// Calculate slash amount & deduct from total
		slashAmount := sdk.NewRatFromInt(unbondingDelegation.InitialBalance.Amount, sdk.OneInt()).Mul(fraction)
		remainingSlashAmount = remainingSlashAmount.Sub(slashAmount)

		// Update unbonding delegation
		slashAmountInt := slashAmount.EvaluateInt()
		if slashAmountInt.GT(unbondingDelegation.Balance.Amount) {
			slashAmountInt = unbondingDelegation.Balance.Amount
		}
		unbondingDelegation.Balance = unbondingDelegation.Balance.Minus(sdk.Coin{unbondingDelegation.Balance.Denom, slashAmountInt})
		k.SetUnbondingDelegation(ctx, unbondingDelegation)
	}

	// Iterate through redelegations from slashed validator
	redelegations := k.GetRedelegationsFromValidator(ctx, address)
	for _, redelegation := range redelegations {
		if redelegation.MinTime < now {
			// TODO Delete element?
			continue
		}

		// Calculate slash amount & deduct from total
		slashAmount := sdk.NewRatFromInt(redelegation.InitialBalance.Amount, sdk.OneInt()).Mul(fraction)
		remainingSlashAmount = remainingSlashAmount.Sub(slashAmount)

		// Update redelegation
		slashAmountInt := slashAmount.EvaluateInt()
		if slashAmountInt.GT(redelegation.Balance.Amount) {
			slashAmountInt = redelegation.Balance.Amount
		}
		redelegation.Balance = redelegation.Balance.Minus(sdk.Coin{redelegation.Balance.Denom, slashAmountInt})
		k.SetRedelegation(ctx, redelegation)
	}

	sharesToRemove := remainingSlashAmount
	// Cannot decrease balance below zero
	if sharesToRemove.GT(validator.PoolShares.Amount) {
		sharesToRemove = validator.PoolShares.Amount
	}

	// Slash the validator & burn tokens
	pool := k.GetPool(ctx)
	validator, pool, burned := validator.RemovePoolShares(pool, sharesToRemove)
	k.SetPool(ctx, pool)              // update the pool
	k.UpdateValidator(ctx, validator) // update the validator, possibly kicking it out

	// Log that a slash occurred!
	logger := ctx.Logger().With("module", "x/stake")
	logger.Info(fmt.Sprintf("Validator %s slashed by fraction %v, removed %v shares and burned %d tokens", pubkey.Address(), fraction, sharesToRemove, burned))

	// TODO Return event(s)
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
	k.UpdateValidator(ctx, validator)
	return
}
