package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/delegation/exported"
	"github.com/cosmos/cosmos-sdk/x/delegation/internal/types"
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
func (k Keeper) DelegateFeeAllowance(ctx sdk.Context, grant types.FeeAllowanceGrant) error {
	store := ctx.KVStore(k.storeKey)
	key := types.FeeAllowanceKey(grant.Granter, grant.Grantee)

	bz, err := k.cdc.MarshalBinaryBare(grant)
	if err != nil {
		return err
	}
	store.Set(key, bz)
	return nil
}

// RevokeFeeAllowance removes an existing grant
func (k Keeper) RevokeFeeAllowance(ctx sdk.Context, granter sdk.AccAddress, grantee sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	key := types.FeeAllowanceKey(granter, grantee)
	store.Delete(key)
}

// GetFeeAllowance returns the allowance between the granter and grantee.
// If there is none, it returns nil, nil.
// Returns an error on parsing issues
func (k Keeper) GetFeeAllowance(ctx sdk.Context, granter sdk.AccAddress, grantee sdk.AccAddress) (exported.FeeAllowance, error) {
	grant, err := k.GetFeeGrant(ctx, granter, grantee)
	if grant == nil {
		return nil, err
	}
	return grant.Allowance, err
}

// GetFeeGrant returns entire grant between both accounts
func (k Keeper) GetFeeGrant(ctx sdk.Context, granter sdk.AccAddress, grantee sdk.AccAddress) (*types.FeeAllowanceGrant, error) {
	store := ctx.KVStore(k.storeKey)
	key := types.FeeAllowanceKey(granter, grantee)
	bz := store.Get(key)
	if len(bz) == 0 {
		return nil, nil
	}

	var grant types.FeeAllowanceGrant
	err := k.cdc.UnmarshalBinaryBare(bz, &grant)
	if err != nil {
		return nil, err
	}
	return &grant, nil
}

// GetAllMyFeeAllowances returns a list of all the grants from anyone to the given grantee.
func (k Keeper) GetAllMyFeeAllowances(ctx sdk.Context, grantee sdk.AccAddress) ([]types.FeeAllowanceGrant, error) {
	store := ctx.KVStore(k.storeKey)
	var grants []types.FeeAllowanceGrant

	prefix := types.FeeAllowancePrefixByGrantee(grantee)
	iter := sdk.KVStorePrefixIterator(store, prefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		bz := iter.Value()
		var grant types.FeeAllowanceGrant
		err := k.cdc.UnmarshalBinaryBare(bz, &grant)
		if err != nil {
			return nil, err
		}
		grants = append(grants, grant)
	}
	return grants, nil
}

// GetAllFeeAllowances returns a list of all the grants in the store.
// This is very expensive and only designed for export genesis
func (k Keeper) GetAllFeeAllowances(ctx sdk.Context) ([]types.FeeAllowanceGrant, error) {
	store := ctx.KVStore(k.storeKey)
	var grants []types.FeeAllowanceGrant

	iter := sdk.KVStorePrefixIterator(store, types.FeeAllowanceKeyPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		bz := iter.Value()
		var grant types.FeeAllowanceGrant
		err := k.cdc.UnmarshalBinaryBare(bz, &grant)
		if err != nil {
			return nil, err
		}
		grants = append(grants, grant)
	}
	return grants, nil
}

// UseDelegatedFees will try to pay the given fee from the granter's account as requested by the grantee
// (true, nil) will update the allowance, and assumes the AnteHandler deducts the given fees
// (false, nil) rejects payment on behalf of grantee
// (?, err) means there was a data parsing error (abort tx and log this info)
func (k Keeper) UseDelegatedFees(ctx sdk.Context, granter sdk.AccAddress, grantee sdk.AccAddress, fee sdk.Coins) bool {
	grant, err := k.GetFeeGrant(ctx, granter, grantee)
	if err != nil {
		// we should acknowledge a db issue somehow (better?)
		ctx.Logger().Error(err.Error())
		return false
	}
	if grant == nil || grant.Allowance == nil {
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
	if err := k.DelegateFeeAllowance(ctx, *grant); err != nil {
		// we should acknowledge a db issue somehow (better?)
		ctx.Logger().Error(err.Error())
		return false
	}
	return true
}
