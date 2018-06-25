package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	crypto "github.com/tendermint/go-crypto"
)

// NOTE the current slash functionality doesn't take into consideration unbonding/rebonding records
//      or the time of breach. This will be updated in slashing v2
// slash a validator
func (k Keeper) Slash(ctx sdk.Context, pubkey crypto.PubKey, height int64, power int64, fraction sdk.Rat) {

	// Amount of slashing = slash fraction * power at time of equivocation
	slashAmount := sdk.NewRat(power).Mul(fraction)
	// hmm, https://github.com/cosmos/cosmos-sdk/issues/1348

	validator, found := k.GetValidatorByPubKey(ctx, pubkey)
	if !found {
		panic(fmt.Errorf("Attempted to slash a nonexistent validator with address %s", pubkey.Address()))
	}

	// Track remaining slash amount
	remainingSlashAmount := slashAmount

	// TODO Iterate through unbondings

	// TODO Iterate through redelegations

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
