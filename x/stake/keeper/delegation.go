package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/tags"
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
func (k PrivlegedKeeper) GetAllDelegations(ctx sdk.Context) (delegations []types.Delegation) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, DelegationKey)

	i := 0
	for ; ; i++ {
		if !iterator.Valid() {
			iterator.Close()
			break
		}
		bondBytes := iterator.Value()
		var delegation types.Delegation
		k.cdc.MustUnmarshalBinary(bondBytes, &delegation)
		delegations = append(delegations, delegation)
		iterator.Next()
	}
	return delegations[:i] // trim
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
			iterator.Close()
			break
		}
		bondBytes := iterator.Value()
		var delegation types.Delegation
		k.cdc.MustUnmarshalBinary(bondBytes, &delegation)
		delegations[i] = delegation
		iterator.Next()
	}
	return delegations[:i] // trim
}

// set the delegation
func (k PrivlegedKeeper) SetDelegation(ctx sdk.Context, delegation types.Delegation) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinary(delegation)
	store.Set(GetDelegationKey(delegation.DelegatorAddr, delegation.ValidatorAddr, k.cdc), b)
}

// remove the delegation
func (k PrivlegedKeeper) RemoveDelegation(ctx sdk.Context, delegation types.Delegation) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(GetDelegationKey(delegation.DelegatorAddr, delegation.ValidatorAddr, k.cdc))
}

// common functionality between handlers
func (k PrivlegedKeeper) Delegate(ctx sdk.Context, delegatorAddr sdk.Address,
	bondAmt sdk.Coin, validator types.Validator) (sdk.Tags, sdk.Error) {

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
	_, _, err := k.coinKeeper.SubtractCoins(ctx, delegation.DelegatorAddr, sdk.Coins{bondAmt})
	if err != nil {
		return nil, err
	}
	validator, pool, newShares := validator.AddTokensFromDel(pool, bondAmt.Amount)
	delegation.Shares = delegation.Shares.Add(newShares)

	// Update delegation height
	delegation.Height = ctx.BlockHeight()

	k.SetPool(ctx, pool)
	k.SetDelegation(ctx, delegation)
	k.UpdateValidator(ctx, validator)
	tags := sdk.NewTags(
		tags.Action, tags.ActionDelegate,
		tags.Delegator, delegatorAddr.Bytes(),
		tags.DstValidator, validator.Owner.Bytes(),
	)
	return tags, nil
}

//_____________________________________________________________________________________

// load a unbonding delegation
func (k Keeper) GetUnbondingDelegation(ctx sdk.Context,
	DelegatorAddr, ValidatorAddr sdk.Address) (ubd types.UnbondingDelegation, found bool) {

	store := ctx.KVStore(k.storeKey)
	ubdKey := GetUBDKey(ubd.DelegatorAddr, ubd.ValidatorAddr, k.cdc)
	bz := store.Get(ubdKey)
	if bz == nil {
		return ubd, false
	}

	k.cdc.MustUnmarshalBinary(bz, &ubd)
	return ubd, true
}

// set the unbonding delegation and associated index
func (k PrivlegedKeeper) SetUnbondingDelegation(ctx sdk.Context, ubd types.UnbondingDelegation) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinary(ubd)
	ubdKey := GetUBDKey(ubd.DelegatorAddr, ubd.ValidatorAddr, k.cdc)
	store.Set(ubdKey, bz)
	store.Set(GetUBDByValIndexKey(ubd.DelegatorAddr, ubd.ValidatorAddr, k.cdc), ubdKey)
}

// remove the unbonding delegation object and associated index
func (k PrivlegedKeeper) RemoveUnbondingDelegation(ctx sdk.Context, ubd types.UnbondingDelegation) {
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
	redKey := GetREDKey(red.DelegatorAddr, red.ValidatorSrcAddr, red.ValidatorDstAddr, k.cdc)
	bz := store.Get(redKey)
	if bz == nil {
		return red, false
	}

	k.cdc.MustUnmarshalBinary(bz, &red)
	return red, true
}

// load a unbonding delegation and the associated delegation
func (k Keeper) GetRedelegationDel(ctx sdk.Context, DelegatorAddr, ValidatorSrcAddr,
	ValidatorDstAddr sdk.Address) (red types.Redelegation, srcDelegation, dstDelegation types.Delegation, found bool) {

	red, found = k.GetRedelegation(ctx, DelegatorAddr, ValidatorSrcAddr, ValidatorDstAddr)
	if !found {
		return red, srcDelegation, dstDelegation, false
	}
	srcDelegation, found = k.GetDelegation(ctx, red.DelegatorAddr, red.ValidatorSrcAddr)
	if !found {
		panic("found redelegation but not source delegation object")
	}
	dstDelegation, found = k.GetDelegation(ctx, red.DelegatorAddr, red.ValidatorDstAddr)
	if !found {
		panic("found redelegation but not source delegation object")
	}
	return red, srcDelegation, dstDelegation, true
}

// set a redelegation and associated index
func (k PrivlegedKeeper) SetRedelegation(ctx sdk.Context, red types.Redelegation) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinary(red)
	redKey := GetREDKey(red.DelegatorAddr, red.ValidatorSrcAddr, red.ValidatorDstAddr, k.cdc)
	store.Set(redKey, bz)
	store.Set(GetREDByValSrcIndexKey(red.DelegatorAddr, red.ValidatorSrcAddr, red.ValidatorDstAddr, k.cdc), redKey)
	store.Set(GetREDByValDstIndexKey(red.DelegatorAddr, red.ValidatorSrcAddr, red.ValidatorDstAddr, k.cdc), redKey)
}

// remove a redelegation object and associated index
func (k PrivlegedKeeper) RemoveRedelegation(ctx sdk.Context, red types.Redelegation) {
	store := ctx.KVStore(k.storeKey)
	redKey := GetREDKey(red.DelegatorAddr, red.ValidatorSrcAddr, red.ValidatorDstAddr, k.cdc)
	store.Delete(redKey)
	store.Delete(GetREDByValSrcIndexKey(red.DelegatorAddr, red.ValidatorSrcAddr, red.ValidatorDstAddr, k.cdc))
	store.Delete(GetREDByValDstIndexKey(red.DelegatorAddr, red.ValidatorSrcAddr, red.ValidatorDstAddr, k.cdc))
}
