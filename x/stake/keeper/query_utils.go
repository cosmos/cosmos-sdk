package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

// Return all validators that a delegator is bonded to. If maxRetrieve is supplied, the respective amount will be returned.
func (k Keeper) GetDelegatorValidators(ctx sdk.Context, delegatorAddr sdk.AccAddress,
	maxRetrieve ...int16) (validators []types.Validator) {

	retrieve := len(maxRetrieve) > 0
	if retrieve {
		validators = make([]types.Validator, maxRetrieve[0])
	}
	store := ctx.KVStore(k.storeKey)
	delegatorPrefixKey := GetDelegationsKey(delegatorAddr)
	iterator := sdk.KVStorePrefixIterator(store, delegatorPrefixKey) //smallest to largest

	i := 0
	for ; iterator.Valid() && (!retrieve || (retrieve && i < int(maxRetrieve[0]))); iterator.Next() {
		addr := iterator.Key()
		delegation := types.MustUnmarshalDelegation(k.cdc, addr, iterator.Value())
		validator, found := k.GetValidator(ctx, delegation.ValidatorAddr)
		if !found {
			panic(types.ErrNoValidatorFound(types.DefaultCodespace))
		}

		if retrieve {
			validators[i] = validator
		} else {
			validators = append(validators, validator)
		}

		i++
	}
	iterator.Close()
	return validators[:i] // trim
}

// Return all validators that a delegator is bonded to. If maxRetrieve is supplied, the respective amount will be returned.
func (k Keeper) GetDelegatorBechValidators(ctx sdk.Context, delegatorAddr sdk.AccAddress,
	maxRetrieve ...int16) (validators []types.BechValidator) {

	retrieve := len(maxRetrieve) > 0
	if retrieve {
		validators = make([]types.BechValidator, maxRetrieve[0])
	}
	store := ctx.KVStore(k.storeKey)
	delegatorPrefixKey := GetDelegationsKey(delegatorAddr)
	iterator := sdk.KVStorePrefixIterator(store, delegatorPrefixKey) //smallest to largest

	i := 0
	for ; iterator.Valid() && (!retrieve || (retrieve && i < int(maxRetrieve[0]))); iterator.Next() {
		addr := iterator.Key()
		delegation := types.MustUnmarshalDelegation(k.cdc, addr, iterator.Value())
		validator, found := k.GetValidator(ctx, delegation.ValidatorAddr)
		if !found {
			panic(types.ErrNoValidatorFound(types.DefaultCodespace))
		}

		bechValidator, err := validator.Bech32Validator()
		if err != nil {
			panic(err.Error())
		}

		if retrieve {
			validators[i] = bechValidator
		} else {
			validators = append(validators, bechValidator)
		}
		i++
	}
	iterator.Close()
	return validators[:i] // trim
}

// return a validator that a delegator is bonded to
func (k Keeper) GetDelegatorValidator(ctx sdk.Context, delegatorAddr sdk.AccAddress,
	validatorAddr sdk.ValAddress) (validator types.Validator) {

	delegation, found := k.GetDelegation(ctx, delegatorAddr, validatorAddr)
	if !found {
		panic(types.ErrNoDelegation(types.DefaultCodespace))
	}

	validator, found = k.GetValidator(ctx, delegation.ValidatorAddr)
	if !found {
		panic(types.ErrNoValidatorFound(types.DefaultCodespace))
	}
	return
}
