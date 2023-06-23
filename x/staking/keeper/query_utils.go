package keeper

import (
	"context"

	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// Return all validators that a delegator is bonded to. If maxRetrieve is supplied, the respective amount will be returned.
func (k Keeper) GetDelegatorValidators(
	ctx context.Context, delegatorAddr sdk.AccAddress, maxRetrieve uint32,
) (types.Validators, error) {
	validators := make([]types.Validator, maxRetrieve)
	store := k.storeService.OpenKVStore(ctx)
	delegatorPrefixKey := types.GetDelegationsKey(delegatorAddr)

	iterator, err := store.Iterator(delegatorPrefixKey, storetypes.PrefixEndBytes(delegatorPrefixKey)) // smallest to largest
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	i := 0
	for ; iterator.Valid() && i < int(maxRetrieve); iterator.Next() {
		delegation := types.MustUnmarshalDelegation(k.cdc, iterator.Value())

		validator, err := k.GetValidator(ctx, delegation.GetValidatorAddr())
		if err != nil {
			return nil, err
		}

		validators[i] = validator
		i++
	}

	return validators[:i], nil // trim
}

// return a validator that a delegator is bonded to
func (k Keeper) GetDelegatorValidator(
	ctx context.Context, delegatorAddr sdk.AccAddress, validatorAddr sdk.ValAddress,
) (validator types.Validator, err error) {
	delegation, err := k.GetDelegation(ctx, delegatorAddr, validatorAddr)
	if err != nil {
		return validator, err
	}

	return k.GetValidator(ctx, delegation.GetValidatorAddr())
}

// return all delegations for a delegator
func (k Keeper) GetAllDelegatorDelegations(ctx context.Context, delegator sdk.AccAddress) ([]types.Delegation, error) {
	delegations := make([]types.Delegation, 0)

	store := k.storeService.OpenKVStore(ctx)
	delegatorPrefixKey := types.GetDelegationsKey(delegator)

	iterator, err := store.Iterator(delegatorPrefixKey, storetypes.PrefixEndBytes(delegatorPrefixKey)) // smallest to largest
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	for i := 0; iterator.Valid(); iterator.Next() {
		delegation, err := types.UnmarshalDelegation(k.cdc, iterator.Value())
		if err != nil {
			return nil, err
		}
		delegations = append(delegations, delegation)
		i++
	}

	return delegations, nil
}

// return all unbonding-delegations for a delegator
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

// return all redelegations for a delegator
func (k Keeper) GetAllRedelegations(
	ctx context.Context, delegator sdk.AccAddress, srcValAddress, dstValAddress sdk.ValAddress,
) ([]types.Redelegation, error) {
	store := k.storeService.OpenKVStore(ctx)
	delegatorPrefixKey := types.GetREDsKey(delegator)

	iterator, err := store.Iterator(delegatorPrefixKey, storetypes.PrefixEndBytes(delegatorPrefixKey)) // smallest to largest
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	srcValFilter := !(srcValAddress.Empty())
	dstValFilter := !(dstValAddress.Empty())

	redelegations := []types.Redelegation{}

	for ; iterator.Valid(); iterator.Next() {
		redelegation := types.MustUnmarshalRED(k.cdc, iterator.Value())
		valSrcAddr, err := sdk.ValAddressFromBech32(redelegation.ValidatorSrcAddress)
		if err != nil {
			return nil, err
		}
		valDstAddr, err := sdk.ValAddressFromBech32(redelegation.ValidatorDstAddress)
		if err != nil {
			return nil, err
		}
		if srcValFilter && !(srcValAddress.Equals(valSrcAddr)) {
			continue
		}

		if dstValFilter && !(dstValAddress.Equals(valDstAddr)) {
			continue
		}

		redelegations = append(redelegations, redelegation)
	}

	return redelegations, nil
}
