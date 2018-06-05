package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

// load a delegator bond
func (k Keeper) GetDelegation(ctx sdk.Context,
	delegatorAddr, validatorAddr sdk.Address) (bond types.Delegation, found bool) {

	store := ctx.KVStore(k.storeKey)
	delegatorBytes := store.Get(GetDelegationKey(delegatorAddr, validatorAddr, k.cdc))
	if delegatorBytes == nil {
		return bond, false
	}

	k.cdc.MustUnmarshalBinary(delegatorBytes, &bond)
	return bond, true
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

// load all bonds of a delegator
func (k Keeper) GetDelegations(ctx sdk.Context, delegator sdk.Address, maxRetrieve int16) (bonds []types.Delegation) {
	store := ctx.KVStore(k.storeKey)
	delegatorPrefixKey := GetDelegationsKey(delegator, k.cdc)
	iterator := sdk.KVStorePrefixIterator(store, delegatorPrefixKey) //smallest to largest

	bonds = make([]types.Delegation, maxRetrieve)
	i := 0
	for ; ; i++ {
		if !iterator.Valid() || i > int(maxRetrieve-1) {
			iterator.Close()
			break
		}
		bondBytes := iterator.Value()
		var bond types.Delegation
		k.cdc.MustUnmarshalBinary(bondBytes, &bond)
		bonds[i] = bond
		iterator.Next()
	}
	return bonds[:i] // trim
}

// set the delegation
func (k PrivlegedKeeper) SetDelegation(ctx sdk.Context, bond types.Delegation) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinary(bond)
	store.Set(GetDelegationKey(bond.DelegatorAddr, bond.ValidatorAddr, k.cdc), b)
}

// remove the delegation
func (k PrivlegedKeeper) RemoveDelegation(ctx sdk.Context, bond types.Delegation) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(GetDelegationKey(bond.DelegatorAddr, bond.ValidatorAddr, k.cdc))
}

// common functionality between handlers
func (k PrivlegedKeeper) Delegate(ctx sdk.Context, delegatorAddr sdk.Address,
	bondAmt sdk.Coin, validator types.Validator) (sdk.Tags, sdk.Error) {

	// Get or create the delegator bond
	bond, found := k.GetDelegation(ctx, delegatorAddr, validator.Owner)
	if !found {
		bond = types.Delegation{
			DelegatorAddr: delegatorAddr,
			ValidatorAddr: validator.Owner,
			Shares:        sdk.ZeroRat(),
		}
	}

	// Account new shares, save
	pool := k.GetPool(ctx)
	_, _, err := k.coinKeeper.SubtractCoins(ctx, bond.DelegatorAddr, sdk.Coins{bondAmt})
	if err != nil {
		return nil, err
	}
	validator, pool, newShares := validator.AddTokensFromDel(pool, bondAmt.Amount)
	bond.Shares = bond.Shares.Add(newShares)

	// Update bond height
	bond.Height = ctx.BlockHeight()

	k.SetPool(ctx, pool)
	k.SetDelegation(ctx, bond)
	k.UpdateValidator(ctx, validator)
	tags := sdk.NewTags(
		"action", []byte("delegate"),
		"delegator", delegatorAddr.Bytes(),
		"validator", validator.Owner.Bytes(),
	)
	return tags, nil
}

//_____________________________________________________________________________________

// load a delegator bond
func (k Keeper) GetUnbondingDelegation(ctx sdk.Context,
	DelegatorAddr, ValidatorAddr sdk.Address) (bond types.UnbondingDelegation, found bool) {

	store := ctx.KVStore(k.storeKey)
	delegatorBytes := store.Get(UnbondingDelegationKey)
	if delegatorBytes == nil {
		return bond, false
	}

	k.cdc.MustUnmarshalBinary(delegatorBytes, &bond)
	return bond, true
}

// set the delegation
func (k PrivlegedKeeper) SetUnbondingDelegation(ctx sdk.Context, bond types.UnbondingDelegation) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinary(bond)
	store.Set(GetDelegationKey(bond.DelegatorAddr, bond.ValidatorAddr, k.cdc), b)
}
