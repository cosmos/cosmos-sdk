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
	defer iterator.Close()

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
	defer iterator.Close()

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

// Return all delegations for a delegator. If maxRetrieve is supplied, the respective amount will be returned.
func (k Keeper) GetDelegatorDelegationsREST(ctx sdk.Context, delegator sdk.AccAddress,
	maxRetrieve ...int16) (delegations []types.DelegationREST) {
	retrieve := len(maxRetrieve) > 0
	if retrieve {
		delegations = make([]types.DelegationREST, maxRetrieve[0])
	}
	store := ctx.KVStore(k.storeKey)
	delegatorPrefixKey := GetDelegationsKey(delegator)
	iterator := sdk.KVStorePrefixIterator(store, delegatorPrefixKey) //smallest to largest
	defer iterator.Close()

	i := 0
	for ; iterator.Valid() && (!retrieve || (retrieve && i < int(maxRetrieve[0]))); iterator.Next() {
		delegation := types.MustUnmarshalDelegation(k.cdc, iterator.Key(), iterator.Value())
		if retrieve {
			delegations[i] = delegation.ToRest()
		} else {
			delegations = append(delegations, delegation.ToRest())
		}
		i++
	}
	return delegations[:i] // trim
}

// Return all redelegations for a delegator. If maxRetrieve is supplied, the respective amount will be returned.
func (k Keeper) GetRedelegationsREST(ctx sdk.Context, delegator sdk.AccAddress,
	maxRetrieve ...int16) (redelegations []types.RedelegationREST) {

	retrieve := len(maxRetrieve) > 0
	if retrieve {
		redelegations = make([]types.RedelegationREST, maxRetrieve[0])
	}
	store := ctx.KVStore(k.storeKey)
	delegatorPrefixKey := GetREDsKey(delegator)
	iterator := sdk.KVStorePrefixIterator(store, delegatorPrefixKey) //smallest to largest
	defer iterator.Close()

	i := 0
	for ; iterator.Valid() && (!retrieve || (retrieve && i < int(maxRetrieve[0]))); iterator.Next() {
		redelegation := types.MustUnmarshalRED(k.cdc, iterator.Key(), iterator.Value())
		if retrieve {
			redelegations[i] = redelegation.ToRest()
		} else {
			redelegations = append(redelegations, redelegation.ToRest())
		}
		i++
	}
	return redelegations[:i] // trim
}

// Get the set of all validators. If maxRetrieve is supplied, the respective amount will be returned.
func (k Keeper) GetBechValidators(ctx sdk.Context, maxRetrieve ...int16) (validators []types.BechValidator) {
	retrieve := len(maxRetrieve) > 0
	if retrieve {
		validators = make([]types.BechValidator, maxRetrieve[0])
	}

	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, ValidatorsKey)
	defer iterator.Close()

	i := 0
	for ; iterator.Valid() && (!retrieve || (retrieve && i < int(maxRetrieve[0]))); iterator.Next() {
		addr := iterator.Key()[1:]
		validator := types.MustUnmarshalValidator(k.cdc, addr, iterator.Value())
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
	return validators[:i] // trim
}
