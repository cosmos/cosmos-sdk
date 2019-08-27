package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/poa/exported"
	"github.com/cosmos/cosmos-sdk/x/poa/internal/types"
)

// iterate through the validator set and perform the provided function
func (k Keeper) IterateValidators(ctx sdk.Context, fn func(index int64, validator exported.ValidatorI) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.ValidatorsKey)
	defer iterator.Close()
	i := int64(0)
	for ; iterator.Valid(); iterator.Next() {
		validator := types.MustUnmarshalValidator(k.cdc, iterator.Value())
		stop := fn(i, validator) // XXX is this safe will the validator unexposed fields be able to get written to?
		if stop {
			break
		}
		i++
	}
}

// iterate through the active validator set and perform the provided function
func (k Keeper) IterateLastValidators(ctx sdk.Context, fn func(index int64, validator exported.ValidatorI) (stop bool)) {
	iterator := k.LastValidatorsIterator(ctx)
	defer iterator.Close()
	i := int64(0)
	for ; iterator.Valid(); iterator.Next() {
		address := types.AddressFromLastValidatorPowerKey(iterator.Key())
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
}

// iterate through the bonded validator set and perform the provided function
func (k Keeper) IterateBondedValidatorsByPower(ctx sdk.Context, fn func(index int64, validator exported.ValidatorI) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	maxValidators := k.MaxValidators(ctx)

	iterator := sdk.KVStoreReversePrefixIterator(store, types.ValidatorsByPowerIndexKey)
	defer iterator.Close()

	i := int64(0)
	for ; iterator.Valid() && i < int64(maxValidators); iterator.Next() {
		address := iterator.Value()
		validator := k.mustGetValidator(ctx, address)

		if validator.IsBonded() {
			stop := fn(i, validator) // XXX is this safe will the validator unexposed fields be able to get written to?
			if stop {
				break
			}
			i++
		}
	}
}

// Validator gets the Validator interface for a particular address
func (k Keeper) Validator(ctx sdk.Context, address sdk.ValAddress) exported.ValidatorI {
	val, found := k.GetValidator(ctx, address)
	if !found {
		return nil
	}
	return val
}

// ValidatorByConsAddr gets the validator interface for a particular pubkey
func (k Keeper) ValidatorByConsAddr(ctx sdk.Context, addr sdk.ConsAddress) exported.ValidatorI {
	val, found := k.GetValidatorByConsAddr(ctx, addr)
	if !found {
		return nil
	}
	return val
}
