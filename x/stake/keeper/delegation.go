package keeper

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

// load a delegation
func (k Keeper) GetDelegation(ctx sdk.Context,
	delegatorAddr, validatorAddr sdk.Address) (delegation types.Delegation, found bool) {

	store := ctx.KVStore(k.storeKey)
	delegatorBytes := store.Get(GetDelegationKey(delegatorAddr, validatorAddr, k.cdc))
	if delegatorBytes == nil {
		return delegation, false
	}

	k.cdc.MustUnmarshalBinary(delegatorBytes, &delegation)
	return delegation, true
}

// load all delegations used during genesis dump
func (k Keeper) GetAllDelegations(ctx sdk.Context) (delegations []types.Delegation) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, DelegationKey)

	i := 0
	for ; ; i++ {
		if !iterator.Valid() {
			break
		}
		bondBytes := iterator.Value()
		var delegation types.Delegation
		k.cdc.MustUnmarshalBinary(bondBytes, &delegation)
		delegations = append(delegations, delegation)
		iterator.Next()
	}
	iterator.Close()
	return delegations
}

// load all delegations for a delegator
func (k Keeper) GetDelegations(ctx sdk.Context, delegator sdk.Address,
	maxRetrieve int16) (delegations []types.Delegation) {

	store := ctx.KVStore(k.storeKey)
	delegatorPrefixKey := GetDelegationsKey(delegator, k.cdc)
	iterator := sdk.KVStorePrefixIterator(store, delegatorPrefixKey) //smallest to largest

	delegations = make([]types.Delegation, maxRetrieve)
	i := 0
	for ; ; i++ {
		if !iterator.Valid() || i > int(maxRetrieve-1) {
			break
		}
		bondBytes := iterator.Value()
		var delegation types.Delegation
		k.cdc.MustUnmarshalBinary(bondBytes, &delegation)
		delegations[i] = delegation
		iterator.Next()
	}
	iterator.Close()
	return delegations[:i] // trim
}

// set the delegation
func (k Keeper) SetDelegation(ctx sdk.Context, delegation types.Delegation) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinary(delegation)
	store.Set(GetDelegationKey(delegation.DelegatorAddr, delegation.ValidatorAddr, k.cdc), b)
}

// remove the delegation
func (k Keeper) RemoveDelegation(ctx sdk.Context, delegation types.Delegation) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(GetDelegationKey(delegation.DelegatorAddr, delegation.ValidatorAddr, k.cdc))
}

//_____________________________________________________________________________________

// load a unbonding delegation
func (k Keeper) GetUnbondingDelegation(ctx sdk.Context,
	DelegatorAddr, ValidatorAddr sdk.Address) (ubd types.UnbondingDelegation, found bool) {

	store := ctx.KVStore(k.storeKey)
	ubdKey := GetUBDKey(DelegatorAddr, ValidatorAddr, k.cdc)
	bz := store.Get(ubdKey)
	if bz == nil {
		return ubd, false
	}

	k.cdc.MustUnmarshalBinary(bz, &ubd)
	return ubd, true
}

// load all unbonding delegations from a particular validator
func (k Keeper) GetUnbondingDelegationsFromValidator(ctx sdk.Context, valAddr sdk.Address) (unbondingDelegations []types.UnbondingDelegation) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, GetUBDsByValIndexKey(valAddr, k.cdc))
	i := 0
	for ; ; i++ {
		if !iterator.Valid() {
			break
		}
		unbondingBytes := iterator.Value()
		var unbondingDelegation types.UnbondingDelegation
		k.cdc.MustUnmarshalBinary(unbondingBytes, &unbondingDelegation)
		unbondingDelegations = append(unbondingDelegations, unbondingDelegation)
		iterator.Next()
	}
	iterator.Close()
	return unbondingDelegations
}

// set the unbonding delegation and associated index
func (k Keeper) SetUnbondingDelegation(ctx sdk.Context, ubd types.UnbondingDelegation) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinary(ubd)
	ubdKey := GetUBDKey(ubd.DelegatorAddr, ubd.ValidatorAddr, k.cdc)
	store.Set(ubdKey, bz)
	store.Set(GetUBDByValIndexKey(ubd.DelegatorAddr, ubd.ValidatorAddr, k.cdc), ubdKey)
}

// remove the unbonding delegation object and associated index
func (k Keeper) RemoveUnbondingDelegation(ctx sdk.Context, ubd types.UnbondingDelegation) {
	store := ctx.KVStore(k.storeKey)
	ubdKey := GetUBDKey(ubd.DelegatorAddr, ubd.ValidatorAddr, k.cdc)
	store.Delete(ubdKey)
	store.Delete(GetUBDByValIndexKey(ubd.DelegatorAddr, ubd.ValidatorAddr, k.cdc))
}

//_____________________________________________________________________________________

// load a redelegation
func (k Keeper) GetRedelegation(ctx sdk.Context,
	DelegatorAddr, ValidatorSrcAddr, ValidatorDstAddr sdk.Address) (red types.Redelegation, found bool) {

	store := ctx.KVStore(k.storeKey)
	redKey := GetREDKey(DelegatorAddr, ValidatorSrcAddr, ValidatorDstAddr, k.cdc)
	bz := store.Get(redKey)
	if bz == nil {
		return red, false
	}

	k.cdc.MustUnmarshalBinary(bz, &red)
	return red, true
}

// load all redelegations from a particular validator
func (k Keeper) GetRedelegationsFromValidator(ctx sdk.Context, valAddr sdk.Address) (redelegations []types.Redelegation) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, GetREDsByValSrcIndexKey(valAddr, k.cdc))
	i := 0
	for ; ; i++ {
		if !iterator.Valid() {
			break
		}
		redelegationBytes := iterator.Value()
		var redelegation types.Redelegation
		k.cdc.MustUnmarshalBinary(redelegationBytes, &redelegation)
		redelegations = append(redelegations, redelegation)
		iterator.Next()
	}
	iterator.Close()
	return redelegations
}

// set a redelegation and associated index
func (k Keeper) SetRedelegation(ctx sdk.Context, red types.Redelegation) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinary(red)
	redKey := GetREDKey(red.DelegatorAddr, red.ValidatorSrcAddr, red.ValidatorDstAddr, k.cdc)
	store.Set(redKey, bz)
	store.Set(GetREDByValSrcIndexKey(red.DelegatorAddr, red.ValidatorSrcAddr, red.ValidatorDstAddr, k.cdc), redKey)
	store.Set(GetREDByValDstIndexKey(red.DelegatorAddr, red.ValidatorSrcAddr, red.ValidatorDstAddr, k.cdc), redKey)
}

// remove a redelegation object and associated index
func (k Keeper) RemoveRedelegation(ctx sdk.Context, red types.Redelegation) {
	store := ctx.KVStore(k.storeKey)
	redKey := GetREDKey(red.DelegatorAddr, red.ValidatorSrcAddr, red.ValidatorDstAddr, k.cdc)
	store.Delete(redKey)
	store.Delete(GetREDByValSrcIndexKey(red.DelegatorAddr, red.ValidatorSrcAddr, red.ValidatorDstAddr, k.cdc))
	store.Delete(GetREDByValDstIndexKey(red.DelegatorAddr, red.ValidatorSrcAddr, red.ValidatorDstAddr, k.cdc))
}

//_____________________________________________________________________________________

// Perform a delegation, set/update everything necessary within the store
func (k Keeper) Delegate(ctx sdk.Context, delegatorAddr sdk.Address, bondAmt sdk.Coin,
	validator types.Validator) (newShares sdk.Rat, err sdk.Error) {

	// Get or create the delegator delegation
	delegation, found := k.GetDelegation(ctx, delegatorAddr, validator.Owner)
	if !found {
		delegation = types.Delegation{
			DelegatorAddr: delegatorAddr,
			ValidatorAddr: validator.Owner,
			Shares:        sdk.ZeroRat(),
		}
	}

	// Account new shares, save
	pool := k.GetPool(ctx)
	_, _, err = k.coinKeeper.SubtractCoins(ctx, delegation.DelegatorAddr, sdk.Coins{bondAmt})
	if err != nil {
		return
	}
	validator, pool, newShares = validator.AddTokensFromDel(pool, bondAmt.Amount.Int64())
	delegation.Shares = delegation.Shares.Add(newShares)

	// Update delegation height
	delegation.Height = ctx.BlockHeight()

	k.SetPool(ctx, pool)
	k.SetDelegation(ctx, delegation)
	k.UpdateValidator(ctx, validator)

	return
}

// unbond the the delegation return
func (k Keeper) Unbond(ctx sdk.Context, delegatorAddr, validatorAddr sdk.Address,
	shares sdk.Rat) (amount int64, err sdk.Error) {

	// check if delegation has any shares in it unbond
	delegation, found := k.GetDelegation(ctx, delegatorAddr, validatorAddr)
	if !found {
		err = types.ErrNoDelegatorForAddress(k.Codespace())
		return
	}

	// retrieve the amount to remove
	if delegation.Shares.LT(shares) {
		err = types.ErrNotEnoughDelegationShares(k.Codespace(), delegation.Shares.String())
		return
	}

	// get validator
	validator, found := k.GetValidator(ctx, validatorAddr)
	if !found {
		err = types.ErrNoValidatorFound(k.Codespace())
		return
	}

	// subtract shares from delegator
	delegation.Shares = delegation.Shares.Sub(shares)

	// remove the delegation
	if delegation.Shares.IsZero() {

		// if the delegation is the owner of the validator then
		// trigger a revoke validator
		if bytes.Equal(delegation.DelegatorAddr, validator.Owner) && validator.Revoked == false {
			validator.Revoked = true
		}
		k.RemoveDelegation(ctx, delegation)
	} else {
		// Update height
		delegation.Height = ctx.BlockHeight()
		k.SetDelegation(ctx, delegation)
	}

	// remove the coins from the validator
	pool := k.GetPool(ctx)
	validator, pool, amount = validator.RemoveDelShares(pool, shares)

	k.SetPool(ctx, pool)

	// update then remove validator if necessary
	validator = k.UpdateValidator(ctx, validator)
	if validator.DelegatorShares.IsZero() {
		k.RemoveValidator(ctx, validator.Owner)
	}

	return
}
