package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

// Implements ValidatorSet
var _ sdk.ValidatorSet = Keeper{}

// iterate through the active validator set and perform the provided function
func (k Keeper) IterateValidators(ctx sdk.Context, fn func(index int64, validator sdk.Validator) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, ValidatorsKey)
	i := int64(0)
	for ; iterator.Valid(); iterator.Next() {
		addr := iterator.Key()[1:]
		validator := types.MustUnmarshalValidator(k.cdc, addr, iterator.Value())
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
	iterator := sdk.KVStorePrefixIterator(store, ValidatorsBondedIndexKey)
	i := int64(0)
	for ; iterator.Valid(); iterator.Next() {
		address := GetAddressFromValBondedIndexKey(iterator.Key())
		validator, found := k.GetValidator(ctx, address)
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
func (k Keeper) Validator(ctx sdk.Context, address sdk.AccAddress) sdk.Validator {
	val, found := k.GetValidator(ctx, address)
	if !found {
		return nil
	}
	return val
}

// get the sdk.validator for a particular pubkey
func (k Keeper) ValidatorByPubKey(ctx sdk.Context, pubkey crypto.PubKey) sdk.Validator {
	val, found := k.GetValidatorByPubKey(ctx, pubkey)
	if !found {
		return nil
	}
	return val
}

// total power from the bond
func (k Keeper) TotalPower(ctx sdk.Context) sdk.Rat {
	pool := k.GetPool(ctx)
	return pool.BondedTokens
}

//__________________________________________________________________________

// Implements DelegationSet

var _ sdk.DelegationSet = Keeper{}

// Returns self as it is both a validatorset and delegationset
func (k Keeper) GetValidatorSet() sdk.ValidatorSet {
	return k
}

// get the delegation for a particular set of delegator and validator addresses
func (k Keeper) Delegation(ctx sdk.Context, addrDel sdk.AccAddress, addrVal sdk.AccAddress) sdk.Delegation {
	bond, ok := k.GetDelegation(ctx, addrDel, addrVal)
	if !ok {
		return nil
	}
	return bond
}

// iterate through the active validator set and perform the provided function
func (k Keeper) IterateDelegations(ctx sdk.Context, delAddr sdk.AccAddress, fn func(index int64, delegation sdk.Delegation) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	key := GetDelegationsKey(delAddr)
	iterator := sdk.KVStorePrefixIterator(store, key)
	i := int64(0)
	for ; iterator.Valid(); iterator.Next() {
		delegation := types.MustUnmarshalDelegation(k.cdc, iterator.Key(), iterator.Value())
		stop := fn(i, delegation) // XXX is this safe will the fields be able to get written to?
		if stop {
			break
		}
		i++
	}
	iterator.Close()
}
