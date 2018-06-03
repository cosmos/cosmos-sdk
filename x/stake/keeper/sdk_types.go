package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

// Implements ValidatorSet
var _ sdk.ValidatorSet = Keeper{}
var _ sdk.SlashValidatorSet = PrivlegedKeeper{}

// iterate through the active validator set and perform the provided function
func (k Keeper) IterateValidators(ctx sdk.Context, fn func(index int64, validator sdk.Validator) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, ValidatorsKey)
	i := int64(0)
	for ; iterator.Valid(); iterator.Next() {
		bz := iterator.Value()
		var validator types.Validator
		k.cdc.MustUnmarshalBinary(bz, &validator)
		stop := fn(i, validator) // XXX is this safe will the validator unexposed fields be able to get written to?
		if stop {
			break
		}
		i++
	}
	iterator.Close()
}

// iterate through the active validator set and perform the provided function
func (k Keeper) IterateValidatorsBonded(ctx sdk.Context, fn func(index int64, validator sdk.Validator) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, ValidatorsBondedKey)
	i := int64(0)
	for ; iterator.Valid(); iterator.Next() {
		address := iterator.Value()
		validator, found := k.getValidator(store, address)
		if !found {
			panic(fmt.Sprintf("validator record not found for address: %v\n", address))
		}

		stop := fn(i, validator) // XXX is this safe will the validator unexposed fields be able to get written to?
		if stop {
			break
		}
		i++
	}
	iterator.Close()
}

// get the sdk.validator for a particular address
func (k Keeper) Validator(ctx sdk.Context, addr sdk.Address) sdk.Validator {
	val, found := k.GetValidator(ctx, addr)
	if !found {
		return nil
	}
	return val
}

// total power from the bond
func (k Keeper) TotalPower(ctx sdk.Context) sdk.Rat {
	pool := k.GetPool(ctx)
	return pool.BondedShares
}

//__________________________________________________________________________

// Implements DelegationSet

var _ sdk.DelegationSet = Keeper{}

// get the delegation for a particular set of delegator and validator addresses
func (k Keeper) Delegation(ctx sdk.Context, addrDel sdk.Address, addrVal sdk.Address) sdk.Delegation {
	bond, ok := k.GetDelegation(ctx, addrDel, addrVal)
	if !ok {
		return nil
	}
	return bond
}

// iterate through the active validator set and perform the provided function
func (k Keeper) IterateDelegators(ctx sdk.Context, delAddr sdk.Address, fn func(index int64, delegation sdk.Delegation) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	key := GetDelegationsKey(delAddr, k.cdc)
	iterator := sdk.KVStorePrefixIterator(store, key)
	i := int64(0)
	for ; iterator.Valid(); iterator.Next() {
		bz := iterator.Value()
		var delegation types.Delegation
		k.cdc.MustUnmarshalBinary(bz, &delegation)
		stop := fn(i, delegation) // XXX is this safe will the fields be able to get written to?
		if stop {
			break
		}
		i++
	}
	iterator.Close()
}
