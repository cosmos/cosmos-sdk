package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/fee_grant/exported"
	"github.com/cosmos/cosmos-sdk/x/fee_grant/internal/types"
)

type Keeper struct {
	cdc      *codec.Codec
	storeKey sdk.StoreKey
}

// NewKeeper creates a DelegationKeeper
func NewKeeper(cdc *codec.Codec, storeKey sdk.StoreKey) Keeper {
	return Keeper{cdc: cdc, storeKey: storeKey}
}

// DelegateFeeAllowance creates a new grant
func (k Keeper) DelegateFeeAllowance(ctx sdk.Context, grant types.FeeAllowanceGrant) {
	store := ctx.KVStore(k.storeKey)
	key := types.FeeAllowanceKey(grant.Granter, grant.Grantee)
	bz := k.cdc.MustMarshalBinaryBare(grant)
	store.Set(key, bz)
}

// RevokeFeeAllowance removes an existing grant
func (k Keeper) RevokeFeeAllowance(ctx sdk.Context, granter, grantee sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	key := types.FeeAllowanceKey(granter, grantee)
	store.Delete(key)
}

// GetFeeAllowance returns the allowance between the granter and grantee.
// If there is none, it returns nil, nil.
// Returns an error on parsing issues
func (k Keeper) GetFeeAllowance(ctx sdk.Context, granter, grantee sdk.AccAddress) (exported.FeeAllowance, error) {
	grant, err := k.GetFeeGrant(ctx, granter, grantee)
	if err != nil {
		return nil, err
	}
	return grant.Allowance, nil
}

// GetFeeGrant returns entire grant between both accounts
func (k Keeper) GetFeeGrant(ctx sdk.Context, granter sdk.AccAddress, grantee sdk.AccAddress) (types.FeeAllowanceGrant, error) {
	store := ctx.KVStore(k.storeKey)
	key := types.FeeAllowanceKey(granter, grantee)
	bz := store.Get(key)
	if len(bz) == 0 {
		return types.FeeAllowanceGrant{}, nil
	}

	var grant types.FeeAllowanceGrant
	err := k.cdc.UnmarshalBinaryBare(bz, &grant)
	if err != nil {
		return types.FeeAllowanceGrant{}, err
	}
	return grant, nil
}

// IterateAllGranteeFeeAllowances iterates over all the grants from anyone to the given grantee.
// Callback to get all data, returns true to stop, false to keep reading
func (k Keeper) IterateAllGranteeFeeAllowances(ctx sdk.Context, grantee sdk.AccAddress, cb func(types.FeeAllowanceGrant) bool) error {
	store := ctx.KVStore(k.storeKey)
	prefix := types.FeeAllowancePrefixByGrantee(grantee)
	iter := sdk.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	stop := false
	for ; iter.Valid() && !stop; iter.Next() {
		bz := iter.Value()
		var grant types.FeeAllowanceGrant
		err := k.cdc.UnmarshalBinaryBare(bz, &grant)
		if err != nil {
			return err
		}
		stop = cb(grant)
	}
	return nil
}

// IterateAllFeeAllowances iterates over all the grants in the store.
// Callback to get all data, returns true to stop, false to keep reading
// Calling this without pagination is very expensive and only designed for export genesis
func (k Keeper) IterateAllFeeAllowances(ctx sdk.Context, cb func(types.FeeAllowanceGrant) bool) error {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.FeeAllowanceKeyPrefix)
	defer iter.Close()

	stop := false
	for ; iter.Valid() && !stop; iter.Next() {
		bz := iter.Value()
		var grant types.FeeAllowanceGrant
		err := k.cdc.UnmarshalBinaryBare(bz, &grant)
		if err != nil {
			return err
		}
		stop = cb(grant)
	}
	return nil
}

// UseDelegatedFees will try to pay the given fee from the granter's account as requested by the grantee
func (k Keeper) UseDelegatedFees(ctx sdk.Context, granter, grantee sdk.AccAddress, fee sdk.Coins) bool {
	grant, err := k.GetFeeGrant(ctx, granter, grantee)
	if err != nil {
		panic(err)
	}
	if grant.Allowance == nil {
		return false
	}

	remove, err := grant.Allowance.Accept(fee, ctx.BlockTime(), ctx.BlockHeight())
	if remove {
		k.RevokeFeeAllowance(ctx, granter, grantee)
		return err == nil
	}
	if err != nil {
		return false
	}

	// if we accepted, store the updated state of the allowance
	k.DelegateFeeAllowance(ctx, grant)
	return true
}
