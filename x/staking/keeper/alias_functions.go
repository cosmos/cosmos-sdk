package keeper

import (
	"context"

	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// Validator Set

// IterateValidators iterates through the validator set and perform the provided function
func (k Keeper) IterateValidators(ctx context.Context, fn func(index int64, validator types.ValidatorI) (stop bool)) error {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.ValidatorsKey, storetypes.PrefixEndBytes(types.ValidatorsKey))
	if err != nil {
		return err
	}
	defer iterator.Close()

	i := int64(0)

	for ; iterator.Valid(); iterator.Next() {
		validator, err := types.UnmarshalValidator(k.cdc, iterator.Value())
		if err != nil {
			return err
		}
		stop := fn(i, validator) // XXX is this safe will the validator unexposed fields be able to get written to?

		if stop {
			break
		}
		i++
	}

	return nil
}

// IterateBondedValidatorsByPower iterates through the bonded validator set and perform the provided function
func (k Keeper) IterateBondedValidatorsByPower(ctx context.Context, fn func(index int64, validator types.ValidatorI) (stop bool)) error {
	store := k.storeService.OpenKVStore(ctx)
	maxValidators, err := k.MaxValidators(ctx)
	if err != nil {
		return err
	}

	iterator, err := store.ReverseIterator(types.ValidatorsByPowerIndexKey, storetypes.PrefixEndBytes(types.ValidatorsByPowerIndexKey))
	if err != nil {
		return err
	}
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

	return nil
}

// Validator gets the Validator interface for a particular address
func (k Keeper) Validator(ctx context.Context, address sdk.ValAddress) (types.ValidatorI, error) {
	return k.GetValidator(ctx, address)
}

// ValidatorByConsAddr gets the validator interface for a particular pubkey
func (k Keeper) ValidatorByConsAddr(ctx context.Context, addr sdk.ConsAddress) (types.ValidatorI, error) {
	return k.GetValidatorByConsAddr(ctx, addr)
}

// Delegation Set

// GetValidatorSet returns self as it is both a validatorset and delegationset
func (k Keeper) GetValidatorSet() types.ValidatorSet {
	return k
}

// Delegation gets the delegation interface for a particular set of delegator and validator addresses
func (k Keeper) Delegation(ctx context.Context, addrDel sdk.AccAddress, addrVal sdk.ValAddress) (types.DelegationI, error) {
	return k.Delegations.Get(ctx, collections.Join(addrDel, addrVal))
}

// IterateDelegations iterates through all of the delegations from a delegator
func (k Keeper) IterateDelegations(ctx context.Context, delAddr sdk.AccAddress,
	fn func(index int64, del types.DelegationI) (stop bool),
) error {
	var i int64
	rng := collections.NewPrefixedPairRange[sdk.AccAddress, sdk.ValAddress](delAddr)
	return k.Delegations.Walk(ctx, rng, func(key collections.Pair[sdk.AccAddress, sdk.ValAddress], del types.Delegation) (stop bool, err error) {
		stop = fn(i, del)
		if stop {
			return true, nil
		}
		i++

		return false, nil
	})
}

// GetAllSDKDelegations returns all delegations used during genesis dump
// TODO: remove this func, change all usage for iterate functionality
func (k Keeper) GetAllSDKDelegations(ctx context.Context) (delegations []types.Delegation, err error) {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.DelegationKey, storetypes.PrefixEndBytes(types.DelegationKey))
	if err != nil {
		return delegations, err
	}
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		delegation, err := types.UnmarshalDelegation(k.cdc, iterator.Value())
		if err != nil {
			return delegations, err
		}
		delegations = append(delegations, delegation)
	}

	return
}
