package keeper

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	"cosmossdk.io/x/staking/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// GetDelegation returns a specific delegation.
func (k Keeper) GetDelegation(ctx sdk.Context,
	delAddr sdk.AccAddress, valAddr sdk.ValAddress,
) (delegation types.Delegation, found bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetDelegationKey(delAddr, valAddr)

	value := store.Get(key)
	if value == nil {
		return delegation, false
	}

	return delegations, nil
}

// IterateAllDelegations iterates through all of the delegations.
func (k Keeper) IterateAllDelegations(ctx sdk.Context, cb func(delegation types.Delegation) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

// IterateDelegatorDelegations iterates through one delegator's delegations
func (k Keeper) IterateDelegatorDelegations(ctx sdk.Context, delegator sdk.AccAddress,
	cb func(delegation types.Delegation) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	delegatorPrefixKey := types.GetDelegationsKey(delegator)
	iterator := sdk.KVStorePrefixIterator(store, delegatorPrefixKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		delegation := types.MustUnmarshalDelegation(k.cdc, iterator.Value())
		if cb(delegation) {
			break
		}
	}
}

// GetAllDelegations returns all delegations used during genesis dump.
func (k Keeper) GetAllDelegations(ctx sdk.Context) (delegations []types.Delegation) {
	k.IterateAllDelegations(ctx, func(delegation types.Delegation) bool {
		delegations = append(delegations, delegation)

		return false, nil
	})
	if err != nil {
		return nil, err
	}

	return delegations, nil
}

// GetValidatorDelegations returns all delegations to a specific validator.
// Useful for querier.
func (k Keeper) GetValidatorDelegations(ctx sdk.Context, valAddr sdk.ValAddress) (delegations []types.Delegation) { //nolint:interfacer
	store := ctx.KVStore(k.storeKey)

	var i uint16
	rng := collections.NewPrefixedPairRange[sdk.AccAddress, sdk.ValAddress](delegator)
	err := k.Delegations.Walk(ctx, rng, func(key collections.Pair[sdk.AccAddress, sdk.ValAddress], del types.Delegation) (stop bool, err error) {
		if i >= maxRetrieve {
			return true, nil
		}
		delegations[i] = del
		i++

		return false, nil
	})
	if err != nil {
		return nil, err
	}

	return delegations[:i], nil // trim if the array length < maxRetrieve
}

// GetDelegatorDelegations returns a given amount of all the delegations from a
// delegator.
func (k Keeper) GetDelegatorDelegations(ctx sdk.Context, delegator sdk.AccAddress,
	maxRetrieve uint16,
) (delegations []types.Delegation) {
	delegations = make([]types.Delegation, maxRetrieve)

	i := 0
	if i < int(maxRetrieve) {
		k.IterateDelegatorDelegations(ctx, delegator, func(delegation types.Delegation) bool {
			delegations[i] = delegation
			i++
			return i >= int(maxRetrieve)
		})
	}
	return delegations[:i] // trim if the array length < maxRetrieve
}

// SetDelegation sets a delegation.
func (k Keeper) SetDelegation(ctx sdk.Context, delegation types.Delegation) {
	delegatorAddress := sdk.MustAccAddressFromBech32(delegation.DelegatorAddress)

	store := ctx.KVStore(k.storeKey)
	b := types.MustMarshalDelegation(k.cdc, delegation)
	store.Set(types.GetDelegationKey(delegatorAddress, delegation.GetValidatorAddr()), b)
}

// RemoveDelegation removes a delegation.
func (k Keeper) RemoveDelegation(ctx sdk.Context, delegation types.Delegation) {
	delegatorAddress := sdk.MustAccAddressFromBech32(delegation.DelegatorAddress)

	k.BeforeDelegationRemoved(ctx, delegatorAddress, delegation.GetValidatorAddr())
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetDelegationKey(delegatorAddress, delegation.GetValidatorAddr()))
}

// GetUnbondingDelegations returns a given amount of all the delegator unbonding-delegations.
func (k Keeper) GetUnbondingDelegations(ctx sdk.Context, delegator sdk.AccAddress,
	maxRetrieve uint16,
) (unbondingDelegations []types.UnbondingDelegation) {
	unbondingDelegations = make([]types.UnbondingDelegation, maxRetrieve)

	return k.DelegationsByValidator.Remove(ctx, collections.Join(sdk.ValAddress(valAddr), sdk.AccAddress(delegatorAddress)))
}

// GetUnbondingDelegations returns a given amount of all the delegator unbonding-delegations.
func (k Keeper) GetUnbondingDelegations(ctx context.Context, delegator sdk.AccAddress, maxRetrieve uint16) (unbondingDelegations []types.UnbondingDelegation, err error) {
	unbondingDelegations = make([]types.UnbondingDelegation, maxRetrieve)

	i := 0
	rng := collections.NewPrefixedPairRange[[]byte, []byte](delegator)
	err = k.UnbondingDelegations.Walk(
		ctx,
		rng,
		func(key collections.Pair[[]byte, []byte], value types.UnbondingDelegation) (stop bool, err error) {
			unbondingDelegations = append(unbondingDelegations, value)
			i++

			if i >= int(maxRetrieve) {
				return true, nil
			}
			return false, nil
		},
	)
	if err != nil {
		return nil, err
	}

	return unbondingDelegations[:i], nil // trim if the array length < maxRetrieve
}

// GetUnbondingDelegation returns a unbonding delegation.
func (k Keeper) GetUnbondingDelegation(
	ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress,
) (ubd types.UnbondingDelegation, found bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetUBDKey(delAddr, valAddr)
	value := store.Get(key)

	if value == nil {
		return ubd, false
	}
	return ubd, nil
}

// GetUnbondingDelegationsFromValidator returns all unbonding delegations from a
// particular validator.
func (k Keeper) GetUnbondingDelegationsFromValidator(ctx context.Context, valAddr sdk.ValAddress) (ubds []types.UnbondingDelegation, err error) {
	store := k.KVStoreService.OpenKVStore(ctx)
	rng := collections.NewPrefixedPairRange[[]byte, []byte](valAddr)
	err = k.UnbondingDelegationByValIndex.Walk(
		ctx,
		rng,
		func(key collections.Pair[[]byte, []byte], value []byte) (stop bool, err error) {
			valAddr := key.K1()
			delAddr := key.K2()
			ubdkey := types.GetUBDKey(delAddr, valAddr)
			ubdValue, err := store.Get(ubdkey)
			if err != nil {
				return true, err
			}
			unbondingDelegation, err := types.UnmarshalUBD(k.cdc, ubdValue)
			if err != nil {
				return true, err
			}
			ubds = append(ubds, unbondingDelegation)
			return false, nil
		},
	)
	if err != nil {
		return ubds, err
	}
	return ubds, nil
}

// GetDelegatorUnbonding returns the total amount a delegator has unbonding.
func (k Keeper) GetDelegatorUnbonding(ctx context.Context, delegator sdk.AccAddress) (math.Int, error) {
	unbonding := math.ZeroInt()
	rng := collections.NewPrefixedPairRange[[]byte, []byte](delegator)
	err := k.UnbondingDelegations.Walk(
		ctx,
		rng,
		func(key collections.Pair[[]byte, []byte], ubd types.UnbondingDelegation) (stop bool, err error) {
			for _, entry := range ubd.Entries {
				unbonding = unbonding.Add(entry.Balance)
			}
			return false, nil
		},
	)
	if err != nil {
		return unbonding, err
	}
	return unbonding, err
}

// GetUnbondingDelegationsFromValidator returns all unbonding delegations from a
// particular validator.
func (k Keeper) GetUnbondingDelegationsFromValidator(ctx sdk.Context, valAddr sdk.ValAddress) (ubds []types.UnbondingDelegation) {
	store := ctx.KVStore(k.storeKey)

	var iterErr error
	err := k.IterateDelegatorDelegations(ctx, delegator, func(delegation types.Delegation) bool {
		validatorAddr, err := k.validatorAddressCodec.StringToBytes(delegation.ValidatorAddress)
		if err != nil {
			iterErr = err
			return true
		}
		validator, err := k.GetValidator(ctx, validatorAddr)
		if err == nil {
			shares := delegation.Shares
			tokens := validator.TokensFromSharesTruncated(shares)
			bonded = bonded.Add(tokens)
		}
		return false
	})
	if iterErr != nil {
		return bonded.RoundInt(), iterErr
	}
	return bonded.RoundInt(), err
}

// IterateDelegatorDelegations iterates through one delegator's delegations.
func (k Keeper) IterateDelegatorDelegations(ctx context.Context, delegator sdk.AccAddress, cb func(delegation types.Delegation) (stop bool)) error {
	rng := collections.NewPrefixedPairRange[sdk.AccAddress, sdk.ValAddress](delegator)
	err := k.Delegations.Walk(ctx, rng, func(key collections.Pair[sdk.AccAddress, sdk.ValAddress], del types.Delegation) (stop bool, err error) {
		if cb(del) {
			return true, nil
		}

		return false, nil
	})
	if err != nil {
		return err
	}

	return nil
}

// IterateUnbondingDelegations iterates through all of the unbonding delegations.
func (k Keeper) IterateUnbondingDelegations(ctx sdk.Context, fn func(index int64, ubd types.UnbondingDelegation) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.GetUBDsKey(delegator))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		ubd := types.MustUnmarshalUBD(k.cdc, iterator.Value())
		if cb(ubd) {
			break
		}
	}
}

// GetDelegatorUnbonding returns the total amount a delegator has unbonding.
func (k Keeper) GetDelegatorUnbonding(ctx sdk.Context, delegator sdk.AccAddress) sdk.Int {
	unbonding := sdk.ZeroInt()
	k.IterateDelegatorUnbondingDelegations(ctx, delegator, func(ubd types.UnbondingDelegation) bool {
		for _, entry := range ubd.Entries {
			unbonding = unbonding.Add(entry.Balance)
		}
		return false
	})
	return unbonding
}

// IterateDelegatorUnbondingDelegations iterates through a delegator's unbonding delegations.
func (k Keeper) IterateDelegatorUnbondingDelegations(ctx sdk.Context, delegator sdk.AccAddress, cb func(ubd types.UnbondingDelegation) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.GetUBDsKey(delegator))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		ubd := types.MustUnmarshalUBD(k.cdc, iterator.Value())
		if cb(ubd) {
			break
		}
	}
}

// GetDelegatorBonded returs the total amount a delegator has bonded.
func (k Keeper) GetDelegatorBonded(ctx sdk.Context, delegator sdk.AccAddress) sdk.Int {
	bonded := sdk.ZeroDec()

	k.IterateDelegatorDelegations(ctx, delegator, func(delegation types.Delegation) bool {
		validatorAddr, err := sdk.ValAddressFromBech32(delegation.ValidatorAddress)
		if err != nil {
			panic(err) // shouldn't happen
		}
		validator, found := k.GetValidator(ctx, validatorAddr)
		if found {
			shares := delegation.Shares
			tokens := validator.TokensFromSharesTruncated(shares)
			bonded = bonded.Add(tokens)
		}
		return false
	})
	return bonded.RoundInt()
}

// IterateDelegatorDelegations iterates through one delegator's delegations.
func (k Keeper) IterateDelegatorDelegations(ctx sdk.Context, delegator sdk.AccAddress, cb func(delegation types.Delegation) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	delegatorPrefixKey := types.GetDelegationsKey(delegator)
	iterator := sdk.KVStorePrefixIterator(store, delegatorPrefixKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		delegation := types.MustUnmarshalDelegation(k.cdc, iterator.Value())
		if cb(delegation) {
			break
		}
	}
}

// IterateDelegatorRedelegations iterates through one delegator's redelegations.
func (k Keeper) IterateDelegatorRedelegations(ctx sdk.Context, delegator sdk.AccAddress, cb func(red types.Redelegation) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	delegatorPrefixKey := types.GetREDsKey(delegator)

	iterator := sdk.KVStorePrefixIterator(store, delegatorPrefixKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		red := types.MustUnmarshalRED(k.cdc, iterator.Value())
		if cb(red) {
			break
		}
	}
}

// HasMaxUnbondingDelegationEntries - check if unbonding delegation has maximum number of entries.
func (k Keeper) HasMaxUnbondingDelegationEntries(ctx sdk.Context,
	delegatorAddr sdk.AccAddress, validatorAddr sdk.ValAddress,
) bool {
	ubd, found := k.GetUnbondingDelegation(ctx, delegatorAddr, validatorAddr)
	if !found {
		return false
	}

	maxEntries, err := k.MaxEntries(ctx)
	if err != nil {
		return false, err
	}
	return len(ubd.Entries) >= int(maxEntries), nil
}

// SetUnbondingDelegation sets the unbonding delegation and associated index.
func (k Keeper) SetUnbondingDelegation(ctx sdk.Context, ubd types.UnbondingDelegation) {
	delegatorAddress := sdk.MustAccAddressFromBech32(ubd.DelegatorAddress)

	store := ctx.KVStore(k.storeKey)
	bz := types.MustMarshalUBD(k.cdc, ubd)
	addr, err := sdk.ValAddressFromBech32(ubd.ValidatorAddress)
	if err != nil {
		return err
	}

	return k.UnbondingDelegationByValIndex.Set(ctx, collections.Join(valAddr, delAddr), []byte{})
}

// RemoveUnbondingDelegation removes the unbonding delegation object and associated index.
func (k Keeper) RemoveUnbondingDelegation(ctx sdk.Context, ubd types.UnbondingDelegation) {
	delegatorAddress := sdk.MustAccAddressFromBech32(ubd.DelegatorAddress)

	store := ctx.KVStore(k.storeKey)
	addr, err := sdk.ValAddressFromBech32(ubd.ValidatorAddress)
	if err != nil {
		return err
	}

	return k.UnbondingDelegationByValIndex.Remove(ctx, collections.Join(valAddr, delAddr))
}

// SetUnbondingDelegationEntry adds an entry to the unbonding delegation at
// the given addresses. It creates the unbonding delegation if it does not exist.
func (k Keeper) SetUnbondingDelegationEntry(
	ctx context.Context, delegatorAddr sdk.AccAddress, validatorAddr sdk.ValAddress,
	creationHeight int64, minTime time.Time, balance math.Int,
) (types.UnbondingDelegation, error) {
	id, err := k.IncrementUnbondingID(ctx)
	if err != nil {
		return types.UnbondingDelegation{}, err
	}

	isNewUbdEntry := true
	ubd, err := k.GetUnbondingDelegation(ctx, delegatorAddr, validatorAddr)
	if err == nil {
		isNewUbdEntry = ubd.AddEntry(creationHeight, minTime, balance, id)
	} else if errors.Is(err, types.ErrNoUnbondingDelegation) {
		ubd = types.NewUnbondingDelegation(delegatorAddr, validatorAddr, creationHeight, minTime, balance, id, k.validatorAddressCodec, k.authKeeper.AddressCodec())
	} else {
		return ubd, err
	}

	if err = k.SetUnbondingDelegation(ctx, ubd); err != nil {
		return ubd, err
	}

	// only call the hook for new entries since
	// calls to AfterUnbondingInitiated are not idempotent
	if isNewUbdEntry {
		// Add to the UBDByUnbondingOp index to look up the UBD by the UBDE ID
		if err = k.SetUnbondingDelegationByUnbondingID(ctx, ubd, id); err != nil {
			return ubd, err
		}

		if err := k.Hooks().AfterUnbondingInitiated(ctx, id); err != nil {
			return ubd, fmt.Errorf("failed to call after unbonding initiated hook: %w", err)
		}
	}
	return ubd, nil
}

// unbonding delegation queue timeslice operations

// GetUBDQueueTimeSlice gets a specific unbonding queue timeslice. A timeslice
// is a slice of DVPairs corresponding to unbonding delegations that expire at a
// certain time.
func (k Keeper) GetUBDQueueTimeSlice(ctx sdk.Context, timestamp time.Time) (dvPairs []types.DVPair) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetUnbondingDelegationTimeKey(timestamp))
	if bz == nil {
		return []types.DVPair{}
	}

	return pairs.Pairs, err
}

// SetUBDQueueTimeSlice sets a specific unbonding queue timeslice.
func (k Keeper) SetUBDQueueTimeSlice(ctx sdk.Context, timestamp time.Time, keys []types.DVPair) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&types.DVPairs{Pairs: keys})
	store.Set(types.GetUnbondingDelegationTimeKey(timestamp), bz)
}

// InsertUBDQueue inserts an unbonding delegation to the appropriate timeslice
// in the unbonding queue.
func (k Keeper) InsertUBDQueue(ctx sdk.Context, ubd types.UnbondingDelegation,
	completionTime time.Time,
) {
	dvPair := types.DVPair{DelegatorAddress: ubd.DelegatorAddress, ValidatorAddress: ubd.ValidatorAddress}

	timeSlice, err := k.GetUBDQueueTimeSlice(ctx, completionTime)
	if err != nil {
		return err
	}
	if len(timeSlice) == 0 {
		if err := k.SetUBDQueueTimeSlice(ctx, completionTime, []types.DVPair{dvPair}); err != nil {
			return err
		}
		return nil
	}

// UBDQueueIterator returns all the unbonding queue timeslices from time 0 until endTime.
func (k Keeper) UBDQueueIterator(ctx sdk.Context, endTime time.Time) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return store.Iterator(types.UnbondingQueueKey,
		sdk.InclusiveEndBytes(types.GetUnbondingDelegationTimeKey(endTime)))
}

// DequeueAllMatureUBDQueue returns a concatenated list of all the timeslices inclusively previous to
// currTime, and deletes the timeslices from the queue.
func (k Keeper) DequeueAllMatureUBDQueue(ctx sdk.Context, currTime time.Time) (matureUnbonds []types.DVPair) {
	store := ctx.KVStore(k.storeKey)

	// gets an iterator for all timeslices from time 0 until the current Blockheader time
	unbondingTimesliceIterator := k.UBDQueueIterator(ctx, ctx.BlockHeader().Time)
	defer unbondingTimesliceIterator.Close()

	for ; iter.Valid(); iter.Next() {
		timeslice, err := iter.Value()
		if err != nil {
			return matureUnbonds, err
		}

		matureUnbonds = append(matureUnbonds, timeslice.Pairs...)
		key, err := iter.Key()
		if err != nil {
			return matureUnbonds, err
		}
		if err = k.UnbondingQueue.Remove(ctx, key); err != nil {
			return matureUnbonds, err
		}
	}

	return matureUnbonds, nil
}

// GetRedelegations returns a given amount of all the delegator redelegations.
func (k Keeper) GetRedelegations(ctx sdk.Context, delegator sdk.AccAddress,
	maxRetrieve uint16,
) (redelegations []types.Redelegation) {
	redelegations = make([]types.Redelegation, maxRetrieve)

	i := 0
	rng := collections.NewPrefixedTripleRange[[]byte, []byte, []byte](delegator)
	err = k.Redelegations.Walk(ctx, rng, func(key collections.Triple[[]byte, []byte, []byte], redelegation types.Redelegation) (stop bool, err error) {
		if i >= int(maxRetrieve) {
			return true, nil
		}

		redelegations[i] = redelegation
		i++
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	return redelegations[:i], nil // trim if the array length < maxRetrieve
}

// GetRedelegation returns a redelegation.
func (k Keeper) GetRedelegation(ctx sdk.Context,
	delAddr sdk.AccAddress, valSrcAddr, valDstAddr sdk.ValAddress,
) (red types.Redelegation, found bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetREDKey(delAddr, valSrcAddr, valDstAddr)

		return false, nil
	})
	if err != nil {
		return nil, err
	}

	return reds, nil
}

// GetRedelegationsFromSrcValidator returns all redelegations from a particular
// validator.
func (k Keeper) GetRedelegationsFromSrcValidator(ctx sdk.Context, valAddr sdk.ValAddress) (reds []types.Redelegation) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.GetREDsFromValSrcIndexKey(valAddr))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		key := types.GetREDKeyFromValSrcIndexKey(iterator.Key())
		value := store.Get(key)
		red := types.MustUnmarshalRED(k.cdc, value)
		reds = append(reds, red)
	}

	return hasReceivingRedelegation, nil
}

// HasReceivingRedelegation checks if validator is receiving a redelegation.
func (k Keeper) HasReceivingRedelegation(ctx sdk.Context,
	delAddr sdk.AccAddress, valDstAddr sdk.ValAddress,
) bool {
	store := ctx.KVStore(k.storeKey)
	prefix := types.GetREDsByDelToValDstIndexKey(delAddr, valDstAddr)

	iterator := sdk.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	return iterator.Valid()
}

// HasMaxRedelegationEntries checks if redelegation has maximum number of entries.
func (k Keeper) HasMaxRedelegationEntries(ctx sdk.Context,
	delegatorAddr sdk.AccAddress, validatorSrcAddr,
	validatorDstAddr sdk.ValAddress,
) bool {
	red, found := k.GetRedelegation(ctx, delegatorAddr, validatorSrcAddr, validatorDstAddr)
	if !found {
		return false
	}

	return len(red.Entries) >= int(maxEntries), nil
}

// SetRedelegation set a redelegation and associated index.
func (k Keeper) SetRedelegation(ctx sdk.Context, red types.Redelegation) {
	delegatorAddress := sdk.MustAccAddressFromBech32(red.DelegatorAddress)

	valSrcAddr, err := k.validatorAddressCodec.StringToBytes(red.ValidatorSrcAddress)
	if err != nil {
		return err
	}
	valDestAddr, err := k.validatorAddressCodec.StringToBytes(red.ValidatorDstAddress)
	if err != nil {
		return err
	}

	if err = k.Redelegations.Set(ctx, collections.Join3(delegatorAddress, valSrcAddr, valDestAddr), red); err != nil {
		return err
	}

	if err = k.RedelegationsByValSrc.Set(ctx, collections.Join3(valSrcAddr, delegatorAddress, valDestAddr), []byte{}); err != nil {
		return err
	}

	return k.RedelegationsByValDst.Set(ctx, collections.Join3(valDestAddr, delegatorAddress, valSrcAddr), []byte{})
}

// SetRedelegationEntry adds an entry to the unbonding delegation at the given
// addresses. It creates the unbonding delegation if it does not exist.
func (k Keeper) SetRedelegationEntry(ctx sdk.Context,
	delegatorAddr sdk.AccAddress, validatorSrcAddr,
	validatorDstAddr sdk.ValAddress, creationHeight int64,
	minTime time.Time, balance sdk.Int,
	sharesSrc, sharesDst sdk.Dec,
) types.Redelegation {
	red, found := k.GetRedelegation(ctx, delegatorAddr, validatorSrcAddr, validatorDstAddr)
	if found {
		red.AddEntry(creationHeight, minTime, balance, sharesDst)
	} else {
		red = types.NewRedelegation(delegatorAddr, validatorSrcAddr,
			validatorDstAddr, creationHeight, minTime, balance, sharesDst, id, k.validatorAddressCodec, k.authKeeper.AddressCodec())
	} else {
		return types.Redelegation{}, err
	}

	if err = k.SetRedelegation(ctx, red); err != nil {
		return types.Redelegation{}, err
	}

	// Add to the UBDByEntry index to look up the UBD by the UBDE ID
	if err = k.SetRedelegationByUnbondingID(ctx, red, id); err != nil {
		return types.Redelegation{}, err
	}

	if err := k.Hooks().AfterUnbondingInitiated(ctx, id); err != nil {
		return types.Redelegation{}, fmt.Errorf("failed to call after unbonding initiated hook: %w", err)
	}

	return red, nil
}

// IterateRedelegations iterates through all redelegations.
func (k Keeper) IterateRedelegations(ctx sdk.Context, fn func(index int64, red types.Redelegation) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	delegatorPrefixKey := types.GetREDsKey(delegator)

	iterator := sdk.KVStorePrefixIterator(store, delegatorPrefixKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		red := types.MustUnmarshalRED(k.cdc, iterator.Value())
		if stop := fn(red); stop {
			break
		}
	}
}

// RemoveRedelegation removes a redelegation object and associated index.
func (k Keeper) RemoveRedelegation(ctx sdk.Context, red types.Redelegation) {
	delegatorAddress := sdk.MustAccAddressFromBech32(red.DelegatorAddress)

	store := ctx.KVStore(k.storeKey)
	valSrcAddr, err := sdk.ValAddressFromBech32(red.ValidatorSrcAddress)
	if err != nil {
		return err
	}
	valDestAddr, err := k.validatorAddressCodec.StringToBytes(red.ValidatorDstAddress)
	if err != nil {
		return err
	}

	if err = k.Redelegations.Remove(ctx, collections.Join3(delegatorAddress, valSrcAddr, valDestAddr)); err != nil {
		return err
	}

	if err = k.RedelegationsByValSrc.Remove(ctx, collections.Join3(valSrcAddr, delegatorAddress, valDestAddr)); err != nil {
		return err
	}

	return k.RedelegationsByValDst.Remove(ctx, collections.Join3(valDestAddr, delegatorAddress, valSrcAddr))
}

// redelegation queue timeslice operations

// GetRedelegationQueueTimeSlice gets a specific redelegation queue timeslice. A
// timeslice is a slice of DVVTriplets corresponding to redelegations that
// expire at a certain time.
func (k Keeper) GetRedelegationQueueTimeSlice(ctx sdk.Context, timestamp time.Time) (dvvTriplets []types.DVVTriplet) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetRedelegationTimeKey(timestamp))
	if bz == nil {
		return []types.DVVTriplet{}
	}

	return triplets.Triplets, nil
}

// SetRedelegationQueueTimeSlice sets a specific redelegation queue timeslice.
func (k Keeper) SetRedelegationQueueTimeSlice(ctx sdk.Context, timestamp time.Time, keys []types.DVVTriplet) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&types.DVVTriplets{Triplets: keys})
	store.Set(types.GetRedelegationTimeKey(timestamp), bz)
}

// InsertRedelegationQueue insert an redelegation delegation to the appropriate
// timeslice in the redelegation queue.
func (k Keeper) InsertRedelegationQueue(ctx sdk.Context, red types.Redelegation,
	completionTime time.Time,
) {
	timeSlice := k.GetRedelegationQueueTimeSlice(ctx, completionTime)
	dvvTriplet := types.DVVTriplet{
		DelegatorAddress:    red.DelegatorAddress,
		ValidatorSrcAddress: red.ValidatorSrcAddress,
		ValidatorDstAddress: red.ValidatorDstAddress,
	}

	if len(timeSlice) == 0 {
		return k.SetRedelegationQueueTimeSlice(ctx, completionTime, []types.DVVTriplet{dvvTriplet})
	}

// RedelegationQueueIterator returns all the redelegation queue timeslices from
// time 0 until endTime.
func (k Keeper) RedelegationQueueIterator(ctx sdk.Context, endTime time.Time) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return store.Iterator(types.RedelegationQueueKey, sdk.InclusiveEndBytes(types.GetRedelegationTimeKey(endTime)))
}

// DequeueAllMatureRedelegationQueue returns a concatenated list of all the
// timeslices inclusively previous to currTime, and deletes the timeslices from
// the queue.
func (k Keeper) DequeueAllMatureRedelegationQueue(ctx sdk.Context, currTime time.Time) (matureRedelegations []types.DVVTriplet) {
	store := ctx.KVStore(k.storeKey)

	// gets an iterator for all timeslices from time 0 until the current Blockheader time
	rng := (&collections.Range[time.Time]{}).EndInclusive(headerInfo.Time)
	err = k.RedelegationQueue.Walk(ctx, rng, func(key time.Time, value types.DVVTriplets) (bool, error) {
		keys = append(keys, key)
		matureRedelegations = append(matureRedelegations, value.Triplets...)
		return false, nil
	})
	if err != nil {
		return matureRedelegations, err
	}
	for _, key := range keys {
		err := k.RedelegationQueue.Remove(ctx, key)
		if err != nil {
			return matureRedelegations, err
		}
	}

	return matureRedelegations, nil
}

// Delegate performs a delegation, set/update everything necessary within the store.
// tokenSrc indicates the bond status of the incoming funds.
func (k Keeper) Delegate(
	ctx context.Context, delAddr sdk.AccAddress, bondAmt math.Int, tokenSrc types.BondStatus,
	validator types.Validator, subtractAccount bool,
) (newShares math.LegacyDec, err error) {
	// In some situations, the exchange rate becomes invalid, e.g. if
	// Validator loses all tokens due to slashing. In this case,
	// make all future delegations invalid.
	if validator.InvalidExRate() {
		return math.LegacyZeroDec(), types.ErrDelegatorShareExRateInvalid
	}

	valbz, err := k.ValidatorAddressCodec().StringToBytes(validator.GetOperator())
	if err != nil {
		return math.LegacyZeroDec(), err
	}

	// Get or create the delegation object and call the appropriate hook if present
	delegation, err := k.Delegations.Get(ctx, collections.Join(delAddr, sdk.ValAddress(valbz)))
	if err == nil {
		// found
		err = k.Hooks().BeforeDelegationSharesModified(ctx, delAddr, valbz)
	} else if errors.Is(err, collections.ErrNotFound) {
		// not found
		delAddrStr, err1 := k.authKeeper.AddressCodec().BytesToString(delAddr)
		if err1 != nil {
			return math.LegacyDec{}, err1
		}

		delegation = types.NewDelegation(delAddrStr, validator.GetOperator(), math.LegacyZeroDec())
		err = k.Hooks().BeforeDelegationCreated(ctx, delAddr, valbz)
	} else {
		return math.LegacyZeroDec(), err
	}

	delegatorAddress := sdk.MustAccAddressFromBech32(delegation.DelegatorAddress)

	// if subtractAccount is true then we are
	// performing a delegation and not a redelegation, thus the source tokens are
	// all non bonded
	if subtractAccount {
		if tokenSrc == types.Bonded {
			return math.LegacyZeroDec(), errors.New("delegation token source cannot be bonded; expected Unbonded or Unbonding, got Bonded")
		}

		var sendName string

		switch {
		case validator.IsBonded():
			sendName = types.BondedPoolName
		case validator.IsUnbonding(), validator.IsUnbonded():
			sendName = types.NotBondedPoolName
		default:
			return math.LegacyZeroDec(), fmt.Errorf("invalid validator status: %v", validator.Status)
		}

		bondDenom, err := k.BondDenom(ctx)
		if err != nil {
			return math.LegacyDec{}, err
		}

		coins := sdk.NewCoins(sdk.NewCoin(bondDenom, bondAmt))
		if err := k.bankKeeper.DelegateCoinsFromAccountToModule(ctx, delAddr, sendName, coins); err != nil {
			return math.LegacyDec{}, err
		}
	} else {
		// potentially transfer tokens between pools, if
		switch {
		case tokenSrc == types.Bonded && validator.IsBonded():
			// do nothing
		case (tokenSrc == types.Unbonded || tokenSrc == types.Unbonding) && !validator.IsBonded():
			// do nothing
		case (tokenSrc == types.Unbonded || tokenSrc == types.Unbonding) && validator.IsBonded():
			// transfer pools
			err = k.notBondedTokensToBonded(ctx, bondAmt)
			if err != nil {
				return math.LegacyDec{}, err
			}
		case tokenSrc == types.Bonded && !validator.IsBonded():
			// transfer pools
			err = k.bondedTokensToNotBonded(ctx, bondAmt)
			if err != nil {
				return math.LegacyDec{}, err
			}
		default:
			return math.LegacyZeroDec(), fmt.Errorf("unknown token source bond status: %v", tokenSrc)
		}
	}

	_, newShares, err = k.AddValidatorTokensAndShares(ctx, validator, bondAmt)
	if err != nil {
		return newShares, err
	}

	// Update delegation
	delegation.Shares = delegation.Shares.Add(newShares)
	if err = k.SetDelegation(ctx, delegation); err != nil {
		return newShares, err
	}

	// Call the after-modification hook
	if err := k.Hooks().AfterDelegationModified(ctx, delAddr, valbz); err != nil {
		return newShares, err
	}

	return newShares, nil
}

// Unbond unbonds a particular delegation and perform associated store operations.
func (k Keeper) Unbond(
	ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress, shares math.LegacyDec,
) (amount math.Int, err error) {
	// check if a delegation object exists in the store
	delegation, err := k.Delegations.Get(ctx, collections.Join(delAddr, valAddr))
	if errors.Is(err, collections.ErrNotFound) {
		return amount, types.ErrNoDelegatorForAddress
	} else if err != nil {
		return amount, err
	}

	// call the before-delegation-modified hook
	if err := k.Hooks().BeforeDelegationSharesModified(ctx, delAddr, valAddr); err != nil {
		return amount, err
	}

	// ensure that we have enough shares to remove
	if delegation.Shares.LT(shares) {
		return amount, errorsmod.Wrap(types.ErrNotEnoughDelegationShares, delegation.Shares.String())
	}

	// get validator
	validator, err := k.GetValidator(ctx, valAddr)
	if err != nil {
		return amount, err
	}

	// subtract shares from delegation
	delegation.Shares = delegation.Shares.Sub(shares)

	delegatorAddress, err := k.authKeeper.AddressCodec().StringToBytes(delegation.DelegatorAddress)
	if err != nil {
		return amount, err
	}

	valbz, err := k.ValidatorAddressCodec().StringToBytes(validator.GetOperator())
	if err != nil {
		return amount, err
	}

	isValidatorOperator := bytes.Equal(delegatorAddress, valbz)

	// If the delegation is the operator of the validator and undelegating will decrease the validator's
	// self-delegation below their minimum, we jail the validator.
	if isValidatorOperator && !validator.Jailed &&
		validator.TokensFromShares(delegation.Shares).TruncateInt().LT(validator.MinSelfDelegation) {
		err = k.jailValidator(ctx, validator)
		if err != nil {
			return amount, fmt.Errorf("failed to jail validator: %w", err)
		}
		validator, err = k.GetValidator(ctx, valbz)
		if err != nil {
			return amount, fmt.Errorf("validator record not found for address: %X", valbz)
		}
	}

	if delegation.Shares.IsZero() {
		err = k.RemoveDelegation(ctx, delegation)
	} else {
		if err = k.SetDelegation(ctx, delegation); err != nil {
			return amount, err
		}

		valAddr, err1 := k.validatorAddressCodec.StringToBytes(delegation.GetValidatorAddr())
		if err1 != nil {
			return amount, err1
		}

		// call the after delegation modification hook
		err = k.Hooks().AfterDelegationModified(ctx, delegatorAddress, valAddr)
	}

	if err != nil {
		return amount, err
	}

	// remove the shares and coins from the validator
	// NOTE that the amount is later (in keeper.Delegation) moved between staking module pools
	validator, amount, err = k.RemoveValidatorTokensAndShares(ctx, validator, shares)
	if err != nil {
		return amount, err
	}

	if validator.DelegatorShares.IsZero() && validator.IsUnbonded() {
		// if not unbonded, we must instead remove validator in EndBlocker once it finishes its unbonding period
		if err = k.RemoveValidator(ctx, valbz); err != nil {
			return amount, err
		}
	}

	return amount, nil
}

// getBeginInfo returns the completion time and height of a redelegation, along
// with a boolean signaling if the redelegation is complete based on the source
// validator.
func (k Keeper) getBeginInfo(
	ctx context.Context, valSrcAddr sdk.ValAddress,
) (completionTime time.Time, height int64, completeNow bool, err error) {
	validator, err := k.GetValidator(ctx, valSrcAddr)
	if err != nil && errors.Is(err, types.ErrNoValidatorFound) {
		return completionTime, height, false, nil
	}
	headerInfo := k.HeaderService.HeaderInfo(ctx)
	unbondingTime, err := k.UnbondingTime(ctx)
	if err != nil {
		return completionTime, height, false, err
	}

	// TODO: When would the validator not be found?
	switch {
	case errors.Is(err, types.ErrNoValidatorFound) || validator.IsBonded():
		// the longest wait - just unbonding period from now
		completionTime = headerInfo.Time.Add(unbondingTime)
		height = headerInfo.Height

		return completionTime, height, false, nil

	case validator.IsUnbonded():
		return completionTime, height, true, nil

	case validator.IsUnbonding():
		return validator.UnbondingTime, validator.UnbondingHeight, false, nil

	default:
		return completionTime, height, false, fmt.Errorf("unknown validator status: %v", validator.Status)
	}
}

// Undelegate unbonds an amount of delegator shares from a given validator. It
// will verify that the unbonding entries between the delegator and validator
// are not exceeded and unbond the staked tokens (based on shares) by creating
// an unbonding object and inserting it into the unbonding queue which will be
// processed during the staking EndBlocker.
func (k Keeper) Undelegate(
	ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress, sharesAmount math.LegacyDec,
) (time.Time, math.Int, error) {
	validator, err := k.GetValidator(ctx, valAddr)
	if err != nil {
		return time.Time{}, math.Int{}, err
	}

	hasMaxEntries, err := k.HasMaxUnbondingDelegationEntries(ctx, delAddr, valAddr)
	if err != nil {
		return time.Time{}, math.Int{}, err
	}

	if hasMaxEntries {
		return time.Time{}, math.Int{}, types.ErrMaxUnbondingDelegationEntries
	}

	returnAmount, err := k.Unbond(ctx, delAddr, valAddr, sharesAmount)
	if err != nil {
		return time.Time{}, math.Int{}, err
	}

	// transfer the validator tokens to the not bonded pool
	if validator.IsBonded() {
		err = k.bondedTokensToNotBonded(ctx, returnAmount)
		if err != nil {
			return time.Time{}, math.Int{}, err
		}
	}

	unbondingTime, err := k.UnbondingTime(ctx)
	if err != nil {
		return time.Time{}, math.Int{}, err
	}

	headerInfo := k.HeaderService.HeaderInfo(ctx)
	completionTime := headerInfo.Time.Add(unbondingTime)
	ubd, err := k.SetUnbondingDelegationEntry(ctx, delAddr, valAddr, headerInfo.Height, completionTime, returnAmount)
	if err != nil {
		return time.Time{}, math.Int{}, err
	}

	err = k.InsertUBDQueue(ctx, ubd, completionTime)
	if err != nil {
		return time.Time{}, math.Int{}, err
	}

	return completionTime, returnAmount, nil
}

// TransferUnbonding changes the ownership of UnbondingDelegation entries
// until the desired number of tokens have changed hands. Returns the actual
// number of tokens transferred.
func (k Keeper) TransferUnbonding(ctx sdk.Context, fromAddr, toAddr sdk.AccAddress, valAddr sdk.ValAddress, wantAmt sdk.Int) sdk.Int {
	transferred := sdk.ZeroInt()
	ubdFrom, found := k.GetUnbondingDelegation(ctx, fromAddr, valAddr)
	if !found {
		return transferred
	}
	ubdFromModified := false

	for i := 0; i < len(ubdFrom.Entries) && wantAmt.IsPositive(); i++ {
		entry := ubdFrom.Entries[i]
		toXfer := entry.Balance
		if toXfer.GT(wantAmt) {
			toXfer = wantAmt
		}
		if !toXfer.IsPositive() {
			continue
		}

		if k.HasMaxUnbondingDelegationEntries(ctx, toAddr, valAddr) {
			// TODO pre-compute the maximum entries we can add rather than checking each time
			break
		}
		ubdTo := k.SetUnbondingDelegationEntry(ctx, toAddr, valAddr, entry.CreationHeight, entry.CompletionTime, toXfer)
		k.InsertUBDQueue(ctx, ubdTo, entry.CompletionTime)
		transferred = transferred.Add(toXfer)
		wantAmt = wantAmt.Sub(toXfer)

		ubdFromModified = true
		remaining := entry.Balance.Sub(toXfer)
		if remaining.IsZero() {
			ubdFrom.RemoveEntry(int64(i))
			i--
			continue
		}
		entry.Balance = remaining
		ubdFrom.Entries[i] = entry
	}

	if ubdFromModified {
		if len(ubdFrom.Entries) == 0 {
			k.RemoveUnbondingDelegation(ctx, ubdFrom)
		} else {
			k.SetUnbondingDelegation(ctx, ubdFrom)
		}
	}
	return transferred
}

// CompleteUnbonding completes the unbonding of all mature entries in the
// retrieved unbonding delegation object and returns the total unbonding balance
// or an error upon failure.
func (k Keeper) CompleteUnbonding(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (sdk.Coins, error) {
	ubd, err := k.GetUnbondingDelegation(ctx, delAddr, valAddr)
	if err != nil {
		return nil, err
	}

	bondDenom, err := k.BondDenom(ctx)
	if err != nil {
		return nil, err
	}

	balances := sdk.NewCoins()
	headerInfo := k.HeaderService.HeaderInfo(ctx)
	ctxTime := headerInfo.Time

	delegatorAddress, err := k.authKeeper.AddressCodec().StringToBytes(ubd.DelegatorAddress)
	if err != nil {
		return nil, err
	}

	// loop through all the entries and complete unbonding mature entries
	for i := 0; i < len(ubd.Entries); i++ {
		entry := ubd.Entries[i]
		if entry.IsMature(ctxTime) && !entry.OnHold() {
			ubd.RemoveEntry(int64(i))
			i--
			if err = k.DeleteUnbondingIndex(ctx, entry.UnbondingId); err != nil {
				return nil, err
			}

			// track undelegation only when remaining or truncated shares are non-zero
			if !entry.Balance.IsZero() {
				amt := sdk.NewCoin(bondDenom, entry.Balance)
				if err := k.bankKeeper.UndelegateCoinsFromModuleToAccount(
					ctx, types.NotBondedPoolName, delegatorAddress, sdk.NewCoins(amt),
				); err != nil {
					return nil, err
				}

				balances = balances.Add(amt)
			}
		}
	}

	// set the unbonding delegation or remove it if there are no more entries
	if len(ubd.Entries) == 0 {
		err = k.RemoveUnbondingDelegation(ctx, ubd)
	} else {
		err = k.SetUnbondingDelegation(ctx, ubd)
	}

	if err != nil {
		return nil, err
	}

	return balances, nil
}

// BeginRedelegation begins unbonding / redelegation and creates a redelegation
// record.
func (k Keeper) BeginRedelegation(
	ctx context.Context, delAddr sdk.AccAddress, valSrcAddr, valDstAddr sdk.ValAddress, sharesAmount math.LegacyDec,
) (completionTime time.Time, err error) {
	if bytes.Equal(valSrcAddr, valDstAddr) {
		return time.Time{}, types.ErrSelfRedelegation
	}

	dstValidator, err := k.GetValidator(ctx, valDstAddr)
	if errors.Is(err, types.ErrNoValidatorFound) {
		return time.Time{}, types.ErrBadRedelegationDst
	} else if err != nil {
		return time.Time{}, err
	}

	srcValidator, err := k.GetValidator(ctx, valSrcAddr)
	if errors.Is(err, types.ErrNoValidatorFound) {
		return time.Time{}, types.ErrBadRedelegationSrc
	} else if err != nil {
		return time.Time{}, err
	}

	// check if this is a transitive redelegation
	hasRecRedel, err := k.HasReceivingRedelegation(ctx, delAddr, valSrcAddr)
	if err != nil {
		return time.Time{}, err
	}

	if hasRecRedel {
		return time.Time{}, types.ErrTransitiveRedelegation
	}

	hasMaxRedels, err := k.HasMaxRedelegationEntries(ctx, delAddr, valSrcAddr, valDstAddr)
	if err != nil {
		return time.Time{}, err
	}

	if hasMaxRedels {
		return time.Time{}, types.ErrMaxRedelegationEntries
	}

	returnAmount, err := k.Unbond(ctx, delAddr, valSrcAddr, sharesAmount)
	if err != nil {
		return time.Time{}, err
	}

	if returnAmount.IsZero() {
		return time.Time{}, types.ErrTinyRedelegationAmount
	}

	sharesCreated, err := k.Delegate(ctx, delAddr, returnAmount, types.BondStatus(srcValidator.GetStatus()), dstValidator, false)
	if err != nil {
		return time.Time{}, err
	}

	// create the unbonding delegation
	completionTime, height, completeNow, err := k.getBeginInfo(ctx, valSrcAddr)
	if err != nil {
		return time.Time{}, err
	}

	if completeNow { // no need to create the redelegation object
		return completionTime, nil
	}

	red, err := k.SetRedelegationEntry(
		ctx, delAddr, valSrcAddr, valDstAddr,
		height, completionTime, returnAmount, sharesAmount, sharesCreated,
	)
	if err != nil {
		return time.Time{}, err
	}

	err = k.InsertRedelegationQueue(ctx, red, completionTime)
	if err != nil {
		return time.Time{}, err
	}

	return completionTime, nil
}

// CompleteRedelegation completes the redelegations of all mature entries in the
// retrieved redelegation object and returns the total redelegation (initial)
// balance or an error upon failure.
func (k Keeper) CompleteRedelegation(
	ctx context.Context, delAddr sdk.AccAddress, valSrcAddr, valDstAddr sdk.ValAddress,
) (sdk.Coins, error) {
	red, err := k.Redelegations.Get(ctx, collections.Join3(delAddr.Bytes(), valSrcAddr.Bytes(), valDstAddr.Bytes()))
	if err != nil {
		return nil, err
	}

	bondDenom, err := k.BondDenom(ctx)
	if err != nil {
		return nil, err
	}

	balances := sdk.NewCoins()
	headerInfo := k.HeaderService.HeaderInfo(ctx)
	ctxTime := headerInfo.Time

	// loop through all the entries and complete mature redelegation entries
	for i := 0; i < len(red.Entries); i++ {
		entry := red.Entries[i]
		if entry.IsMature(ctxTime) && !entry.OnHold() {
			red.RemoveEntry(int64(i))
			i--
			if err = k.DeleteUnbondingIndex(ctx, entry.UnbondingId); err != nil {
				return nil, err
			}

			if !entry.InitialBalance.IsZero() {
				balances = balances.Add(sdk.NewCoin(bondDenom, entry.InitialBalance))
			}
		}
	}

	// set the redelegation or remove it if there are no more entries
	if len(red.Entries) == 0 {
		err = k.RemoveRedelegation(ctx, red)
	} else {
		err = k.SetRedelegation(ctx, red)
	}

	if err != nil {
		return nil, err
	}

	return balances, nil
}

// ValidateUnbondAmount validates that a given unbond or redelegation amount is
// valid based on upon the converted shares. If the amount is valid, the total
// amount of respective shares is returned, otherwise an error is returned.
func (k Keeper) ValidateUnbondAmount(
	ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress, amt math.Int,
) (shares math.LegacyDec, err error) {
	validator, err := k.GetValidator(ctx, valAddr)
	if err != nil {
		return shares, err
	}

	del, err := k.Delegations.Get(ctx, collections.Join(delAddr, valAddr))
	if err != nil {
		return shares, err
	}

	shares, err = validator.SharesFromTokens(amt)
	if err != nil {
		return shares, err
	}

	sharesTruncated, err := validator.SharesFromTokensTruncated(amt)
	if err != nil {
		return shares, err
	}

	delShares := del.GetShares()
	if sharesTruncated.GT(delShares) {
		return shares, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "invalid shares amount")
	}

	// Depending on the share, amount can be smaller than unit amount(1stake).
	// If the remain amount after unbonding is smaller than the minimum share,
	// it's completely unbonded to avoid leaving dust shares.
	tolerance, err := validator.SharesFromTokens(math.OneInt())
	if err != nil {
		return shares, err
	}

	if delShares.Sub(shares).LT(tolerance) {
		shares = delShares
	}

	return shares, nil
}
