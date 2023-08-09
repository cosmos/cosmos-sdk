package keeper

import (
	"context"

	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// GetDelegatorValidators returns all validators that a delegator is bonded to. If maxRetrieve is supplied, the respective amount will be returned.
func (k Keeper) GetDelegatorValidators(
	ctx context.Context, delegatorAddr sdk.AccAddress, maxRetrieve uint32,
) (types.Validators, error) {
	validators := make([]types.Validator, maxRetrieve)

	var i uint32
	rng := collections.NewPrefixedPairRange[sdk.AccAddress, sdk.ValAddress](delegatorAddr)
	err := k.Delegations.Walk(ctx, rng, func(key collections.Pair[sdk.AccAddress, sdk.ValAddress], del types.Delegation) (stop bool, err error) {
		if i >= maxRetrieve {
			return true, nil
		}

		valAddr, err := k.validatorAddressCodec.StringToBytes(del.GetValidatorAddr())
		if err != nil {
			return true, err
		}

		validator, err := k.GetValidator(ctx, valAddr)
		if err != nil {
			return true, err
		}

		validators[i] = validator
		i++

		return false, nil
	})
	if err != nil {
		return nil, err
	}

	return validators[:i], nil // trim
}

// GetDelegatorValidator returns a validator that a delegator is bonded to
func (k Keeper) GetDelegatorValidator(
	ctx context.Context, delegatorAddr sdk.AccAddress, validatorAddr sdk.ValAddress,
) (validator types.Validator, err error) {
	delegation, err := k.Delegations.Get(ctx, collections.Join(delegatorAddr, validatorAddr))
	if err != nil {
		return validator, err
	}

	valAddr, err := k.validatorAddressCodec.StringToBytes(delegation.GetValidatorAddr())
	if err != nil {
		return validator, err
	}

	return k.GetValidator(ctx, valAddr)
}

// GetAllDelegatorDelegations returns all delegations of a delegator
func (k Keeper) GetAllDelegatorDelegations(ctx context.Context, delegator sdk.AccAddress) ([]types.Delegation, error) {
	delegations := make([]types.Delegation, 0)

	var i int64
	rng := collections.NewPrefixedPairRange[sdk.AccAddress, sdk.ValAddress](delegator)
	err := k.Delegations.Walk(ctx, rng, func(key collections.Pair[sdk.AccAddress, sdk.ValAddress], del types.Delegation) (stop bool, err error) {
		delegations = append(delegations, del)
		i++

		return false, nil
	})
	if err != nil {
		return nil, err
	}

	return delegations, nil
}

// GetAllUnbondingDelegations returns all unbonding-delegations of a delegator
func (k Keeper) GetAllUnbondingDelegations(ctx context.Context, delegator sdk.AccAddress) ([]types.UnbondingDelegation, error) {
	unbondingDelegations := make([]types.UnbondingDelegation, 0)

	store := k.storeService.OpenKVStore(ctx)
	delegatorPrefixKey := types.GetUBDsKey(delegator)

	iterator, err := store.Iterator(delegatorPrefixKey, storetypes.PrefixEndBytes(delegatorPrefixKey)) // smallest to largest
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	for i := 0; iterator.Valid(); iterator.Next() {
		unbondingDelegation, err := types.UnmarshalUBD(k.cdc, iterator.Value())
		if err != nil {
			return nil, err
		}
		unbondingDelegations = append(unbondingDelegations, unbondingDelegation)
		i++
	}

	return unbondingDelegations, nil
}

// GetAllRedelegations returns all redelegations of a delegator
func (k Keeper) GetAllRedelegations(
	ctx context.Context, delegator sdk.AccAddress, srcValAddress, dstValAddress sdk.ValAddress,
) ([]types.Redelegation, error) {

	srcValFilter := !(srcValAddress.Empty())
	dstValFilter := !(dstValAddress.Empty())

	redelegations := []types.Redelegation{}
	rng := collections.NewPrefixedTripleRange[sdk.AccAddress, sdk.ValAddress, sdk.ValAddress](delegator)
	err := k.Redelegations.Walk(ctx, rng,
		func(key collections.Triple[sdk.AccAddress, sdk.ValAddress, sdk.ValAddress], redelegation types.Redelegation) (stop bool, err error) {
			valSrcAddr, valDstAddr := key.K2(), key.K3()

			if srcValFilter && !(srcValAddress.Equals(sdk.ValAddress(valSrcAddr))) {
				return false, nil
			}

			if dstValFilter && !(dstValAddress.Equals(sdk.ValAddress(valDstAddr))) {
				return false, nil
			}

			redelegations = append(redelegations, redelegation)
			return false, nil
		},
	)
	if err != nil {
		return nil, err
	}

	return redelegations, nil
}
