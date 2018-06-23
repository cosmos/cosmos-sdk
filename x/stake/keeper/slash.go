package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	crypto "github.com/tendermint/go-crypto"
)

// NOTE the current slash functionality doesn't take into consideration unbonding/rebonding records
//      or the time of breach. This will be updated in slashing v2
// slash a validator
func (k Keeper) Slash(ctx sdk.Context, pubkey crypto.PubKey, height int64, fraction sdk.Rat) {

	// TODO height ignored for now, see https://github.com/cosmos/cosmos-sdk/pull/1011#issuecomment-390253957
	validator, found := k.GetValidatorByPubKey(ctx, pubkey)
	if !found {
		panic(fmt.Errorf("Attempted to slash a nonexistent validator with address %s", pubkey.Address()))
	}
	sharesToRemove := validator.PoolShares.Amount.Mul(fraction)
	pool := k.GetPool(ctx)
	validator, pool, burned := validator.RemovePoolShares(pool, sharesToRemove)
	k.SetPool(ctx, pool)              // update the pool
	k.UpdateValidator(ctx, validator) // update the validator, possibly kicking it out

	logger := ctx.Logger().With("module", "x/stake")
	logger.Info(fmt.Sprintf("Validator %s slashed by fraction %v, removed %v shares and burned %d tokens", pubkey.Address(), fraction, sharesToRemove, burned))
	return
}

// revoke a validator
func (k Keeper) Revoke(ctx sdk.Context, pubkey crypto.PubKey) {

	validator, found := k.GetValidatorByPubKey(ctx, pubkey)
	if !found {
		panic(fmt.Errorf("Validator with pubkey %s not found, cannot revoke", pubkey))
	}
	validator.Revoked = true
	k.UpdateValidator(ctx, validator) // update the validator, now revoked

	logger := ctx.Logger().With("module", "x/stake")
	logger.Info(fmt.Sprintf("Validator %s revoked", pubkey.Address()))
	return
}

// unrevoke a validator
func (k Keeper) Unrevoke(ctx sdk.Context, pubkey crypto.PubKey) {

	validator, found := k.GetValidatorByPubKey(ctx, pubkey)
	if !found {
		panic(fmt.Errorf("Validator with pubkey %s not found, cannot unrevoke", pubkey))
	}
	validator.Revoked = false
	k.UpdateValidator(ctx, validator) // update the validator, now unrevoked

	logger := ctx.Logger().With("module", "x/stake")
	logger.Info(fmt.Sprintf("Validator %s unrevoked", pubkey.Address()))
	return
}
