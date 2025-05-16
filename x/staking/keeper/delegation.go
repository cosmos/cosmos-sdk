package keeper

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	stdmath "math"
	"time"

	corestore "cosmossdk.io/core/store"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// GetDelegation returns a specific delegation.
func (k Keeper) GetDelegation(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (types.Delegation, error) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetDelegationKey(delAddr, valAddr)

	value, err := store.Get(key)
	if err != nil {
		return types.Delegation{}, err
	}

	if value == nil {
		return types.Delegation{}, types.ErrNoDelegation
	}

	return types.UnmarshalDelegation(k.cdc, value)
}

// IterateAllDelegations iterates through all of the delegations.
func (k Keeper) IterateAllDelegations(ctx context.Context, cb func(delegation types.Delegation) (stop bool)) error {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.DelegationKey, storetypes.PrefixEndBytes(types.DelegationKey))
	if err != nil {
		return err
	}
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		delegation := types.MustUnmarshalDelegation(k.cdc, iterator.Value())
		if cb(delegation) {
			break
		}
	}

	return nil
}

// GetAllDelegations returns all delegations used during genesis dump.
func (k Keeper) GetAllDelegations(ctx context.Context) (delegations []types.Delegation, err error) {
	err = k.IterateAllDelegations(ctx, func(delegation types.Delegation) bool {
		delegations = append(delegations, delegation)
		return false
	})

	return delegations, err
}

// GetValidatorDelegations returns all delegations to a specific validator.
// Useful for querier.
func (k Keeper) GetValidatorDelegations(ctx context.Context, valAddr sdk.ValAddress) (delegations []types.Delegation, err error) {
	store := k.storeService.OpenKVStore(ctx)
	prefix := types.GetDelegationsByValPrefixKey(valAddr)
	iterator, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return delegations, err
	}
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var delegation types.Delegation
		valAddr, delAddr, err := types.ParseDelegationsByValKey(iterator.Key())
		if err != nil {
			return delegations, err
		}

		bz, err := store.Get(types.GetDelegationKey(delAddr, valAddr))
		if err != nil {
			return delegations, err
		}

		if err := k.cdc.Unmarshal(bz, &delegation); err != nil {
			return delegations, err
		}

		delegations = append(delegations, delegation)
	}

	return delegations, nil
}

// GetDelegatorDelegations returns a given amount of all the delegations from a
// delegator.
func (k Keeper) GetDelegatorDelegations(ctx context.Context, delegator sdk.AccAddress, maxRetrieve uint16) (delegations []types.Delegation, err error) {
	delegations = make([]types.Delegation, maxRetrieve)
	store := k.storeService.OpenKVStore(ctx)
	delegatorPrefixKey := types.GetDelegationsKey(delegator)

	iterator, err := store.Iterator(delegatorPrefixKey, storetypes.PrefixEndBytes(delegatorPrefixKey))
	if err != nil {
		return delegations, err
	}
	defer iterator.Close()

	i := 0
	for ; iterator.Valid() && i < int(maxRetrieve); iterator.Next() {
		delegation, err := types.UnmarshalDelegation(k.cdc, iterator.Value())
		if err != nil {
			return delegations, err
		}
		delegations[i] = delegation
		i++
	}

	return delegations[:i], nil // trim if the array length < maxRetrieve
}

// SetDelegation sets a delegation.
func (k Keeper) SetDelegation(ctx context.Context, delegation types.Delegation) error {
	delegatorAddress, err := k.authKeeper.AddressCodec().StringToBytes(delegation.DelegatorAddress)
	if err != nil {
		return err
	}

	valAddr, err := k.validatorAddressCodec.StringToBytes(delegation.GetValidatorAddr())
	if err != nil {
		return err
	}

	store := k.storeService.OpenKVStore(ctx)
	b := types.MustMarshalDelegation(k.cdc, delegation)
	err = store.Set(types.GetDelegationKey(delegatorAddress, valAddr), b)
	if err != nil {
		return err
	}

	// set the delegation in validator delegator index
	return store.Set(types.GetDelegationsByValKey(valAddr, delegatorAddress), []byte{})
}

// RemoveDelegation removes a delegation
func (k Keeper) RemoveDelegation(ctx context.Context, delegation types.Delegation) error {
	delegatorAddress, err := k.authKeeper.AddressCodec().StringToBytes(delegation.DelegatorAddress)
	if err != nil {
		return err
	}

	valAddr, err := k.validatorAddressCodec.StringToBytes(delegation.GetValidatorAddr())
	if err != nil {
		return err
	}

	// TODO: Consider calling hooks outside of the store wrapper functions, it's unobvious.
	if err := k.Hooks().BeforeDelegationRemoved(ctx, delegatorAddress, valAddr); err != nil {
		return err
	}

	store := k.storeService.OpenKVStore(ctx)
	err = store.Delete(types.GetDelegationKey(delegatorAddress, valAddr))
	if err != nil {
		return err
	}

	return store.Delete(types.GetDelegationsByValKey(valAddr, delegatorAddress))
}

// GetUnbondingDelegations returns a given amount of all the delegator unbonding-delegations.
func (k Keeper) GetUnbondingDelegations(ctx context.Context, delegator sdk.AccAddress, maxRetrieve uint16) (unbondingDelegations []types.UnbondingDelegation, err error) {
	unbondingDelegations = make([]types.UnbondingDelegation, maxRetrieve)

	store := k.storeService.OpenKVStore(ctx)
	delegatorPrefixKey := types.GetUBDsKey(delegator)

	iterator, err := store.Iterator(delegatorPrefixKey, storetypes.PrefixEndBytes(delegatorPrefixKey))
	if err != nil {
		return unbondingDelegations, err
	}
	defer iterator.Close()

	i := 0
	for ; iterator.Valid() && i < int(maxRetrieve); iterator.Next() {
		unbondingDelegation, err := types.UnmarshalUBD(k.cdc, iterator.Value())
		if err != nil {
			return unbondingDelegations, err
		}
		unbondingDelegations[i] = unbondingDelegation
		i++
	}

	return unbondingDelegations[:i], nil // trim if the array length < maxRetrieve
}

// GetUnbondingDelegation returns a unbonding delegation.
func (k Keeper) GetUnbondingDelegation(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (ubd types.UnbondingDelegation, err error) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetUBDKey(delAddr, valAddr)
	value, err := store.Get(key)
	if err != nil {
		return ubd, err
	}

	if value == nil {
		return ubd, types.ErrNoUnbondingDelegation
	}

	return types.UnmarshalUBD(k.cdc, value)
}

// GetUnbondingDelegationsFromValidator returns all unbonding delegations from a
// particular validator.
func (k Keeper) GetUnbondingDelegationsFromValidator(ctx context.Context, valAddr sdk.ValAddress) (ubds []types.UnbondingDelegation, err error) {
	store := k.storeService.OpenKVStore(ctx)
	prefix := types.GetUBDsByValIndexKey(valAddr)
	iterator, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return ubds, err
	}
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		key := types.GetUBDKeyFromValIndexKey(iterator.Key())
		value, err := store.Get(key)
		if err != nil {
			return ubds, err
		}
		ubd, err := types.UnmarshalUBD(k.cdc, value)
		if err != nil {
			return ubds, err
		}
		ubds = append(ubds, ubd)
	}

	return ubds, nil
}

// IterateUnbondingDelegations iterates through all of the unbonding delegations.
func (k Keeper) IterateUnbondingDelegations(ctx context.Context, fn func(index int64, ubd types.UnbondingDelegation) (stop bool)) error {
	store := k.storeService.OpenKVStore(ctx)
	prefix := types.UnbondingDelegationKey
	iterator, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return err
	}
	defer iterator.Close()

	for i := int64(0); iterator.Valid(); iterator.Next() {
		ubd, err := types.UnmarshalUBD(k.cdc, iterator.Value())
		if err != nil {
			return err
		}
		if stop := fn(i, ubd); stop {
			break
		}
		i++
	}

	return nil
}

// GetDelegatorUnbonding returns the total amount a delegator has unbonding.
func (k Keeper) GetDelegatorUnbonding(ctx context.Context, delegator sdk.AccAddress) (math.Int, error) {
	unbonding := math.ZeroInt()
	err := k.IterateDelegatorUnbondingDelegations(ctx, delegator, func(ubd types.UnbondingDelegation) bool {
		for _, entry := range ubd.Entries {
			unbonding = unbonding.Add(entry.Balance)
		}
		return false
	})
	return unbonding, err
}

// IterateDelegatorUnbondingDelegations iterates through a delegator's unbonding delegations.
func (k Keeper) IterateDelegatorUnbondingDelegations(ctx context.Context, delegator sdk.AccAddress, cb func(ubd types.UnbondingDelegation) (stop bool)) error {
	store := k.storeService.OpenKVStore(ctx)
	prefix := types.GetUBDsKey(delegator)
	iterator, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return err
	}
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		ubd, err := types.UnmarshalUBD(k.cdc, iterator.Value())
		if err != nil {
			return err
		}
		if cb(ubd) {
			break
		}
	}

	return nil
}

// GetDelegatorBonded returs the total amount a delegator has bonded.
func (k Keeper) GetDelegatorBonded(ctx context.Context, delegator sdk.AccAddress) (math.Int, error) {
	bonded := math.LegacyZeroDec()

	err := k.IterateDelegatorDelegations(ctx, delegator, func(delegation types.Delegation) bool {
		validatorAddr, err := k.validatorAddressCodec.StringToBytes(delegation.ValidatorAddress)
		if err != nil {
			panic(err) // shouldn't happen
		}
		validator, err := k.GetValidator(ctx, validatorAddr)
		if err == nil {
			shares := delegation.Shares
			tokens := validator.TokensFromSharesTruncated(shares)
			bonded = bonded.Add(tokens)
		}
		return false
	})
	return bonded.RoundInt(), err
}

// IterateDelegatorDelegations iterates through one delegator's delegations.
func (k Keeper) IterateDelegatorDelegations(ctx context.Context, delegator sdk.AccAddress, cb func(delegation types.Delegation) (stop bool)) error {
	store := k.storeService.OpenKVStore(ctx)
	prefix := types.GetDelegationsKey(delegator)
	iterator, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return err
	}
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		delegation, err := types.UnmarshalDelegation(k.cdc, iterator.Value())
		if err != nil {
			return err
		}
		if cb(delegation) {
			break
		}
	}
	return nil
}

// IterateDelegatorRedelegations iterates through one delegator's redelegations.
func (k Keeper) IterateDelegatorRedelegations(ctx context.Context, delegator sdk.AccAddress, cb func(red types.Redelegation) (stop bool)) error {
	store := k.storeService.OpenKVStore(ctx)
	delegatorPrefixKey := types.GetREDsKey(delegator)
	iterator, err := store.Iterator(delegatorPrefixKey, storetypes.PrefixEndBytes(delegatorPrefixKey))
	if err != nil {
		return err
	}

	for ; iterator.Valid(); iterator.Next() {
		red, err := types.UnmarshalRED(k.cdc, iterator.Value())
		if err != nil {
			return err
		}
		if cb(red) {
			break
		}
	}
	return nil
}

// HasMaxUnbondingDelegationEntries checks if unbonding delegation has maximum number of entries.
func (k Keeper) HasMaxUnbondingDelegationEntries(ctx context.Context, delegatorAddr sdk.AccAddress, validatorAddr sdk.ValAddress) (bool, error) {
	ubd, err := k.GetUnbondingDelegation(ctx, delegatorAddr, validatorAddr)
	if err != nil && !errors.Is(err, types.ErrNoUnbondingDelegation) {
		return false, err
	}

	maxEntries, err := k.MaxEntries(ctx)
	if err != nil {
		return false, err
	}
	return len(ubd.Entries) >= int(maxEntries), nil
}

// SetUnbondingDelegation sets the unbonding delegation and associated index.
func (k Keeper) SetUnbondingDelegation(ctx context.Context, ubd types.UnbondingDelegation) error {
	delAddr, err := k.authKeeper.AddressCodec().StringToBytes(ubd.DelegatorAddress)
	if err != nil {
		return err
	}

	store := k.storeService.OpenKVStore(ctx)
	bz := types.MustMarshalUBD(k.cdc, ubd)
	valAddr, err := k.validatorAddressCodec.StringToBytes(ubd.ValidatorAddress)
	if err != nil {
		return err
	}
	key := types.GetUBDKey(delAddr, valAddr)
	err = store.Set(key, bz)
	if err != nil {
		return err
	}

	return store.Set(types.GetUBDByValIndexKey(delAddr, valAddr), []byte{}) // index, store empty bytes
}

// RemoveUnbondingDelegation removes the unbonding delegation object and associated index.
func (k Keeper) RemoveUnbondingDelegation(ctx context.Context, ubd types.UnbondingDelegation) error {
	delegatorAddress, err := k.authKeeper.AddressCodec().StringToBytes(ubd.DelegatorAddress)
	if err != nil {
		return err
	}

	store := k.storeService.OpenKVStore(ctx)
	addr, err := k.validatorAddressCodec.StringToBytes(ubd.ValidatorAddress)
	if err != nil {
		return err
	}
	key := types.GetUBDKey(delegatorAddress, addr)
	err = store.Delete(key)
	if err != nil {
		return err
	}

	return store.Delete(types.GetUBDByValIndexKey(delegatorAddress, addr))
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
			k.Logger(ctx).Error("failed to call after unbonding initiated hook", "error", err)
		}
	}
	return ubd, nil
}

// unbonding delegation queue timeslice operations

// GetUBDQueueTimeSlice gets a specific unbonding queue timeslice. A timeslice
// is a slice of DVPairs corresponding to unbonding delegations that expire at a
// certain time.
func (k Keeper) GetUBDQueueTimeSlice(ctx context.Context, timestamp time.Time) (dvPairs []types.DVPair, err error) {
	store := k.storeService.OpenKVStore(ctx)

	bz, err := store.Get(types.GetUnbondingDelegationTimeKey(timestamp))
	if bz == nil || err != nil {
		return []types.DVPair{}, err
	}

	pairs := types.DVPairs{}
	err = k.cdc.Unmarshal(bz, &pairs)

	return pairs.Pairs, err
}

// SetUBDQueueTimeSlice sets a specific unbonding queue timeslice.
func (k Keeper) SetUBDQueueTimeSlice(ctx context.Context, timestamp time.Time, keys []types.DVPair) error {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := k.cdc.Marshal(&types.DVPairs{Pairs: keys})
	if err != nil {
		return err
	}
	return store.Set(types.GetUnbondingDelegationTimeKey(timestamp), bz)
}

// InsertUBDQueue inserts an unbonding delegation to the appropriate timeslice
// in the unbonding queue.
func (k Keeper) InsertUBDQueue(ctx context.Context, ubd types.UnbondingDelegation, completionTime time.Time) error {
	dvPair := types.DVPair{DelegatorAddress: ubd.DelegatorAddress, ValidatorAddress: ubd.ValidatorAddress}

	timeSlice, err := k.GetUBDQueueTimeSlice(ctx, completionTime)
	if err != nil {
		return err
	}

	if len(timeSlice) == 0 {
		if err = k.SetUBDQueueTimeSlice(ctx, completionTime, []types.DVPair{dvPair}); err != nil {
			return err
		}
		return nil
	}

	timeSlice = append(timeSlice, dvPair)
	return k.SetUBDQueueTimeSlice(ctx, completionTime, timeSlice)
}

// UBDQueueIterator returns all the unbonding queue timeslices from time 0 until endTime.
func (k Keeper) UBDQueueIterator(ctx context.Context, endTime time.Time) (corestore.Iterator, error) {
	store := k.storeService.OpenKVStore(ctx)
	return store.Iterator(types.UnbondingQueueKey,
		storetypes.InclusiveEndBytes(types.GetUnbondingDelegationTimeKey(endTime)))
}

// DequeueAllMatureUBDQueue returns a concatenated list of all the timeslices inclusively previous to
// currTime, and deletes the timeslices from the queue.
func (k Keeper) DequeueAllMatureUBDQueue(ctx context.Context, currTime time.Time) (matureUnbonds []types.DVPair, err error) {
	store := k.storeService.OpenKVStore(ctx)

	// gets an iterator for all timeslices from time 0 until the current Blockheader time
	unbondingTimesliceIterator, err := k.UBDQueueIterator(ctx, currTime)
	if err != nil {
		return matureUnbonds, err
	}
	defer unbondingTimesliceIterator.Close()

	for ; unbondingTimesliceIterator.Valid(); unbondingTimesliceIterator.Next() {
		timeslice := types.DVPairs{}
		value := unbondingTimesliceIterator.Value()
		if err = k.cdc.Unmarshal(value, &timeslice); err != nil {
			return matureUnbonds, err
		}

		matureUnbonds = append(matureUnbonds, timeslice.Pairs...)

		if err = store.Delete(unbondingTimesliceIterator.Key()); err != nil {
			return matureUnbonds, err
		}

	}

	return matureUnbonds, nil
}

// GetRedelegations returns a given amount of all the delegator redelegations.
func (k Keeper) GetRedelegations(ctx context.Context, delegator sdk.AccAddress, maxRetrieve uint16) (redelegations []types.Redelegation, err error) {
	redelegations = make([]types.Redelegation, maxRetrieve)

	store := k.storeService.OpenKVStore(ctx)
	delegatorPrefixKey := types.GetREDsKey(delegator)
	iterator, err := store.Iterator(delegatorPrefixKey, storetypes.PrefixEndBytes(delegatorPrefixKey))
	if err != nil {
		return nil, err
	}

	i := 0
	for ; iterator.Valid() && i < int(maxRetrieve); iterator.Next() {
		redelegation, err := types.UnmarshalRED(k.cdc, iterator.Value())
		if err != nil {
			return nil, err
		}
		redelegations[i] = redelegation
		i++
	}

	return redelegations[:i], nil // trim if the array length < maxRetrieve
}

// GetRedelegation returns a redelegation.
func (k Keeper) GetRedelegation(ctx context.Context, delAddr sdk.AccAddress, valSrcAddr, valDstAddr sdk.ValAddress) (red types.Redelegation, err error) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetREDKey(delAddr, valSrcAddr, valDstAddr)

	value, err := store.Get(key)
	if err != nil {
		return red, err
	}

	if value == nil {
		return red, types.ErrNoRedelegation
	}

	return types.UnmarshalRED(k.cdc, value)
}

// GetRedelegationsFromSrcValidator returns all redelegations from a particular
// validator.
func (k Keeper) GetRedelegationsFromSrcValidator(ctx context.Context, valAddr sdk.ValAddress) (reds []types.Redelegation, err error) {
	store := k.storeService.OpenKVStore(ctx)
	prefix := types.GetREDsFromValSrcIndexKey(valAddr)
	iterator, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		key := types.GetREDKeyFromValSrcIndexKey(iterator.Key())
		value, err := store.Get(key)
		if err != nil {
			return nil, err
		}
		red, err := types.UnmarshalRED(k.cdc, value)
		if err != nil {
			return nil, err
		}
		reds = append(reds, red)
	}

	return reds, nil
}

// HasReceivingRedelegation checks if validator is receiving a redelegation.
func (k Keeper) HasReceivingRedelegation(ctx context.Context, delAddr sdk.AccAddress, valDstAddr sdk.ValAddress) (bool, error) {
	store := k.storeService.OpenKVStore(ctx)
	prefix := types.GetREDsByDelToValDstIndexKey(delAddr, valDstAddr)
	iterator, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return false, err
	}
	defer iterator.Close()
	return iterator.Valid(), nil
}

// HasMaxRedelegationEntries checks if the redelegation entries reached maximum limit.
func (k Keeper) HasMaxRedelegationEntries(ctx context.Context, delegatorAddr sdk.AccAddress, validatorSrcAddr, validatorDstAddr sdk.ValAddress) (bool, error) {
	red, err := k.GetRedelegation(ctx, delegatorAddr, validatorSrcAddr, validatorDstAddr)
	if err != nil {
		if err == types.ErrNoRedelegation {
			return false, nil
		}

		return false, err
	}
	maxEntries, err := k.MaxEntries(ctx)
	if err != nil {
		return false, err
	}

	return len(red.Entries) >= int(maxEntries), nil
}

// SetRedelegation sets a redelegation and associated index.
func (k Keeper) SetRedelegation(ctx context.Context, red types.Redelegation) error {
	delegatorAddress, err := k.authKeeper.AddressCodec().StringToBytes(red.DelegatorAddress)
	if err != nil {
		return err
	}

	store := k.storeService.OpenKVStore(ctx)
	bz := types.MustMarshalRED(k.cdc, red)
	valSrcAddr, err := k.validatorAddressCodec.StringToBytes(red.ValidatorSrcAddress)
	if err != nil {
		return err
	}
	valDestAddr, err := k.validatorAddressCodec.StringToBytes(red.ValidatorDstAddress)
	if err != nil {
		return err
	}
	key := types.GetREDKey(delegatorAddress, valSrcAddr, valDestAddr)
	if err = store.Set(key, bz); err != nil {
		return err
	}

	if err = store.Set(types.GetREDByValSrcIndexKey(delegatorAddress, valSrcAddr, valDestAddr), []byte{}); err != nil {
		return err
	}

	return store.Set(types.GetREDByValDstIndexKey(delegatorAddress, valSrcAddr, valDestAddr), []byte{})
}

// SetRedelegationEntry adds an entry to the unbonding delegation at the given
// addresses. It creates the unbonding delegation if it does not exist.
func (k Keeper) SetRedelegationEntry(ctx context.Context,
	delegatorAddr sdk.AccAddress, validatorSrcAddr,
	validatorDstAddr sdk.ValAddress, creationHeight int64,
	minTime time.Time, balance math.Int,
	sharesSrc, sharesDst math.LegacyDec,
) (types.Redelegation, error) {
	id, err := k.IncrementUnbondingID(ctx)
	if err != nil {
		return types.Redelegation{}, err
	}

	red, err := k.GetRedelegation(ctx, delegatorAddr, validatorSrcAddr, validatorDstAddr)
	if err == nil {
		red.AddEntry(creationHeight, minTime, balance, sharesDst, id)
	} else if errors.Is(err, types.ErrNoRedelegation) {
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
		k.Logger(ctx).Error("failed to call after unbonding initiated hook", "error", err)
		// TODO (Facu): Should we return here? We are ignoring this error
	}

	return red, nil
}

// IterateRedelegations iterates through all redelegations.
func (k Keeper) IterateRedelegations(ctx context.Context, fn func(index int64, red types.Redelegation) (stop bool)) error {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.RedelegationKey, storetypes.PrefixEndBytes(types.RedelegationKey))
	if err != nil {
		return err
	}
	defer iterator.Close()

	for i := int64(0); iterator.Valid(); iterator.Next() {
		red, err := types.UnmarshalRED(k.cdc, iterator.Value())
		if err != nil {
			return err
		}
		if stop := fn(i, red); stop {
			break
		}
		i++
	}

	return nil
}

// RemoveRedelegation removes a redelegation object and associated index.
func (k Keeper) RemoveRedelegation(ctx context.Context, red types.Redelegation) error {
	delegatorAddress, err := k.authKeeper.AddressCodec().StringToBytes(red.DelegatorAddress)
	if err != nil {
		return err
	}

	store := k.storeService.OpenKVStore(ctx)
	valSrcAddr, err := k.validatorAddressCodec.StringToBytes(red.ValidatorSrcAddress)
	if err != nil {
		return err
	}
	valDestAddr, err := k.validatorAddressCodec.StringToBytes(red.ValidatorDstAddress)
	if err != nil {
		return err
	}
	redKey := types.GetREDKey(delegatorAddress, valSrcAddr, valDestAddr)
	if err = store.Delete(redKey); err != nil {
		return err
	}

	if err = store.Delete(types.GetREDByValSrcIndexKey(delegatorAddress, valSrcAddr, valDestAddr)); err != nil {
		return err
	}

	return store.Delete(types.GetREDByValDstIndexKey(delegatorAddress, valSrcAddr, valDestAddr))
}

// redelegation queue timeslice operations

// GetRedelegationQueueTimeSlice gets a specific redelegation queue timeslice. A
// timeslice is a slice of DVVTriplets corresponding to redelegations that
// expire at a certain time.
func (k Keeper) GetRedelegationQueueTimeSlice(ctx context.Context, timestamp time.Time) (dvvTriplets []types.DVVTriplet, err error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.GetRedelegationTimeKey(timestamp))
	if err != nil {
		return nil, err
	}

	if bz == nil {
		return []types.DVVTriplet{}, nil
	}

	triplets := types.DVVTriplets{}
	err = k.cdc.Unmarshal(bz, &triplets)
	if err != nil {
		return nil, err
	}

	return triplets.Triplets, nil
}

// SetRedelegationQueueTimeSlice sets a specific redelegation queue timeslice.
func (k Keeper) SetRedelegationQueueTimeSlice(ctx context.Context, timestamp time.Time, keys []types.DVVTriplet) error {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := k.cdc.Marshal(&types.DVVTriplets{Triplets: keys})
	if err != nil {
		return err
	}
	return store.Set(types.GetRedelegationTimeKey(timestamp), bz)
}

// InsertRedelegationQueue insert an redelegation delegation to the appropriate
// timeslice in the redelegation queue.
func (k Keeper) InsertRedelegationQueue(ctx context.Context, red types.Redelegation, completionTime time.Time) error {
	timeSlice, err := k.GetRedelegationQueueTimeSlice(ctx, completionTime)
	if err != nil {
		return err
	}
	dvvTriplet := types.DVVTriplet{
		DelegatorAddress:    red.DelegatorAddress,
		ValidatorSrcAddress: red.ValidatorSrcAddress,
		ValidatorDstAddress: red.ValidatorDstAddress,
	}

	if len(timeSlice) == 0 {
		return k.SetRedelegationQueueTimeSlice(ctx, completionTime, []types.DVVTriplet{dvvTriplet})
	}

	timeSlice = append(timeSlice, dvvTriplet)
	return k.SetRedelegationQueueTimeSlice(ctx, completionTime, timeSlice)
}

// RedelegationQueueIterator returns all the redelegation queue timeslices from
// time 0 until endTime.
func (k Keeper) RedelegationQueueIterator(ctx context.Context, endTime time.Time) (storetypes.Iterator, error) {
	store := k.storeService.OpenKVStore(ctx)
	return store.Iterator(types.RedelegationQueueKey, storetypes.InclusiveEndBytes(types.GetRedelegationTimeKey(endTime)))
}

// DequeueAllMatureRedelegationQueue returns a concatenated list of all the
// timeslices inclusively previous to currTime, and deletes the timeslices from
// the queue.
func (k Keeper) DequeueAllMatureRedelegationQueue(ctx context.Context, currTime time.Time) (matureRedelegations []types.DVVTriplet, err error) {
	store := k.storeService.OpenKVStore(ctx)

	// gets an iterator for all timeslices from time 0 until the current Blockheader time
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	redelegationTimesliceIterator, err := k.RedelegationQueueIterator(ctx, sdkCtx.HeaderInfo().Time)
	if err != nil {
		return nil, err
	}
	defer redelegationTimesliceIterator.Close()

	for ; redelegationTimesliceIterator.Valid(); redelegationTimesliceIterator.Next() {
		timeslice := types.DVVTriplets{}
		value := redelegationTimesliceIterator.Value()
		if err = k.cdc.Unmarshal(value, &timeslice); err != nil {
			return nil, err
		}

		matureRedelegations = append(matureRedelegations, timeslice.Triplets...)

		if err = store.Delete(redelegationTimesliceIterator.Key()); err != nil {
			return nil, err
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
	delegation, err := k.GetDelegation(ctx, delAddr, valbz)
	if err == nil {
		// found
		err = k.Hooks().BeforeDelegationSharesModified(ctx, delAddr, valbz)
	} else if errors.Is(err, types.ErrNoDelegation) {
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

	if err != nil {
		return math.LegacyZeroDec(), err
	}

	// if subtractAccount is true then we are
	// performing a delegation and not a redelegation, thus the source tokens are
	// all non bonded
	if subtractAccount {
		if tokenSrc == types.Bonded {
			panic("delegation token source cannot be bonded")
		}

		var sendName string

		switch {
		case validator.IsBonded():
			sendName = types.BondedPoolName
		case validator.IsUnbonding(), validator.IsUnbonded():
			sendName = types.NotBondedPoolName
		default:
			panic("invalid validator status")
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
			panic("unknown token source bond status")
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
	delegation, err := k.GetDelegation(ctx, delAddr, valAddr)
	if errors.Is(err, types.ErrNoDelegation) {
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
			return amount, err
		}
		validator = k.mustGetValidator(ctx, valbz)
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
		return
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	unbondingTime, err := k.UnbondingTime(ctx)
	if err != nil {
		return
	}

	// TODO: When would the validator not be found?
	switch {
	case errors.Is(err, types.ErrNoValidatorFound) || validator.IsBonded():
		// the longest wait - just unbonding period from now
		completionTime = sdkCtx.BlockHeader().Time.Add(unbondingTime)
		height = sdkCtx.BlockHeight()

		return completionTime, height, false, nil

	case validator.IsUnbonded():
		return completionTime, height, true, nil

	case validator.IsUnbonding():
		return validator.UnbondingTime, validator.UnbondingHeight, false, nil

	default:
		panic(fmt.Sprintf("unknown validator status: %s", validator.Status))
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

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	completionTime := sdkCtx.BlockHeader().Time.Add(unbondingTime)
	ubd, err := k.SetUnbondingDelegationEntry(ctx, delAddr, valAddr, sdkCtx.BlockHeight(), completionTime, returnAmount)
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
func (k Keeper) TransferUnbonding(ctx context.Context, fromAddr, toAddr sdk.AccAddress, valAddr sdk.ValAddress, wantAmt math.Int) (math.Int, error) {
	transferred := math.ZeroInt()
	ubdFrom, err := k.GetUnbondingDelegation(ctx, fromAddr, valAddr)
	if errors.Is(err, types.ErrNoUnbondingDelegation) {
		return transferred, nil
	}
	if err != nil {
		return math.Int{}, err
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

		hasMaxUnboundingDelegation, err := k.HasMaxUnbondingDelegationEntries(ctx, toAddr, valAddr)
		if err != nil {
			return math.Int{}, err
		}
		if hasMaxUnboundingDelegation {
			// TODO pre-compute the maximum entries we can add rather than checking each time
			break
		}
		ubdTo, err := k.SetUnbondingDelegationEntry(ctx, toAddr, valAddr, entry.CreationHeight, entry.CompletionTime, toXfer)
		if err != nil {
			return math.Int{}, err
		}
		err = k.InsertUBDQueue(ctx, ubdTo, entry.CompletionTime)
		if err != nil {
			return math.Int{}, err
		}
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
	return transferred, nil
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
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	ctxTime := sdkCtx.BlockHeader().Time

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

	sharesCreated, err := k.Delegate(ctx, delAddr, returnAmount, srcValidator.GetStatus(), dstValidator, false)
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
	red, err := k.GetRedelegation(ctx, delAddr, valSrcAddr, valDstAddr)
	if err != nil {
		return nil, err
	}

	bondDenom, err := k.BondDenom(ctx)
	if err != nil {
		return nil, err
	}

	balances := sdk.NewCoins()
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	ctxTime := sdkCtx.BlockHeader().Time

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

	del, err := k.GetDelegation(ctx, delAddr, valAddr)
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

	// Cap the shares at the delegation's shares. Shares being greater could occur
	// due to rounding, however we don't want to truncate the shares or take the
	// minimum because we want to allow for the full withdraw of shares from a
	// delegation.
	if shares.GT(delShares) {
		shares = delShares
	}

	return shares, nil
}

// TransferDelegation changes the ownership of at most the desired number of shares.
// Returns the actual number of shares transferred. Will also transfer redelegation
// entries to ensure that all redelegations are matched by sufficient shares.
// Note that no tokens are transferred to or from any pool or account, since no
// delegation is actually changing state.
func (k Keeper) TransferDelegation(ctx context.Context, fromAddr, toAddr sdk.AccAddress, valAddr sdk.ValAddress, wantShares math.LegacyDec) (math.LegacyDec, error) {
	transferred := math.LegacyZeroDec()

	// sanity checks
	if !wantShares.IsPositive() {
		return transferred, errors.New("wantShares needs to be positive")
	}
	validator, err := k.GetValidator(ctx, valAddr)
	if err != nil {
		return transferred, err
	}
	delFrom, err := k.GetDelegation(ctx, fromAddr, valAddr)
	if err != nil {
		return transferred, err
	}

	valAddrString, err := k.validatorAddressCodec.BytesToString(valAddr)
	if err != nil {
		return transferred, err
	}

	maxEntries, err := k.MaxEntries(ctx)
	if err != nil {
		return transferred, err
	}

	operatorAddress, err := k.validatorAddressCodec.StringToBytes(validator.GetOperator())
	if err != nil {
		return transferred, err
	}

	toAddrString, err := k.validatorAddressCodec.BytesToString(toAddr)
	if err != nil {
		return transferred, err
	}

	// Check redelegation entry limits while we can still return early.
	// Assume the worst case that we need to transfer all redelegation entries
	mightExceedLimit := false
	k.IterateDelegatorRedelegations(ctx, fromAddr, func(toRedelegation types.Redelegation) (stop bool) {
		// There's no redelegation index by delegator and dstVal or vice-versa.
		// The minimum cardinality is to look up by delegator, so scan and skip.
		if toRedelegation.ValidatorDstAddress != valAddrString {
			return false
		}
		fromRedelegation, err := k.GetRedelegation(ctx, fromAddr, sdk.ValAddress(toRedelegation.ValidatorSrcAddress), sdk.ValAddress(toRedelegation.ValidatorDstAddress))
		if err != nil && len(toRedelegation.Entries)+len(fromRedelegation.Entries) >= int(maxEntries) {
			mightExceedLimit = true
			return true
		}
		return false
	})
	if mightExceedLimit {
		// avoid types.ErrMaxRedelegationEntries
		return transferred, nil
	}

	// compute shares to transfer, amount left behind
	transferred = delFrom.Shares
	if transferred.GT(wantShares) {
		transferred = wantShares
	}
	remaining := delFrom.Shares.Sub(transferred)

	// Update or create the delTo object, calling appropriate hooks
	delTo, err := k.GetDelegation(ctx, toAddr, operatorAddress)
	if err != nil {
		delTo = types.NewDelegation(toAddrString, validator.GetOperator(), math.LegacyZeroDec())
	}
	if err == nil {
		k.Hooks().BeforeDelegationSharesModified(ctx, toAddr, operatorAddress)
	} else {
		k.Hooks().BeforeDelegationCreated(ctx, toAddr, operatorAddress)
	}
	delTo.Shares = delTo.Shares.Add(transferred)
	k.SetDelegation(ctx, delTo)
	k.Hooks().AfterDelegationModified(ctx, toAddr, valAddr)

	// Update source delegation
	if remaining.IsZero() {
		k.Hooks().BeforeDelegationRemoved(ctx, fromAddr, valAddr)
		k.RemoveDelegation(ctx, delFrom)
	} else {
		k.Hooks().BeforeDelegationSharesModified(ctx, fromAddr, valAddr)
		delFrom.Shares = remaining
		k.SetDelegation(ctx, delFrom)
		k.Hooks().AfterDelegationModified(ctx, fromAddr, valAddr)
	}

	// If there are not enough remaining shares to be responsible for
	// the redelegations, transfer some redelegations.
	// For instance, if the original delegation of 300 shares to validator A
	// had redelegations for 100 shares each from validators B, C, and D,
	// and if we're transferring 175 shares, then we might keep the redelegation
	// from B, transfer the one from D, and split the redelegation from C
	// keeping a liability for 25 shares and transferring one for 75 shares.
	// Of course, the redelegations themselves can have multiple entries for
	// different timestamps, so we're actually working at a finer granularity.
	var redelegationErrors []error
	redelegations, err := k.GetRedelegations(ctx, fromAddr, stdmath.MaxUint16)
	if err != nil {
		redelegationErrors = append(redelegationErrors, err)
	}
	for _, redelegation := range redelegations {
		// There's no redelegation index by delegator and dstVal or vice-versa.
		// The minimum cardinality is to look up by delegator, so scan and skip.
		if redelegation.ValidatorDstAddress != valAddrString {
			continue
		}
		redelegationModified := false
		entriesRemaining := false
		for i := 0; i < len(redelegation.Entries); i++ {
			entry := redelegation.Entries[i]

			// Partition SharesDst between keeping and sending
			sharesToKeep := entry.SharesDst
			sharesToSend := math.LegacyZeroDec()
			if entry.SharesDst.GT(remaining) {
				sharesToKeep = remaining
				sharesToSend = entry.SharesDst.Sub(sharesToKeep)
			}
			remaining = remaining.Sub(sharesToKeep) // fewer local shares available to cover liability

			if sharesToSend.IsZero() {
				// Leave the entry here
				entriesRemaining = true
				continue
			}
			if sharesToKeep.IsZero() {
				// Transfer the whole entry, delete locally
				toRed, err := k.SetRedelegationEntry(
					ctx, toAddr, sdk.ValAddress(redelegation.ValidatorSrcAddress),
					sdk.ValAddress(redelegation.ValidatorDstAddress),
					entry.CreationHeight, entry.CompletionTime, entry.InitialBalance, math.LegacyZeroDec(), sharesToSend,
				)
				if err != nil {
					redelegationErrors = append(redelegationErrors, err)
					continue
				}

				k.InsertRedelegationQueue(ctx, toRed, entry.CompletionTime)
				redelegation.RemoveEntry(int64(i))
				i--
				// okay to leave an obsolete entry in the queue for the removed entry
				redelegationModified = true
			} else {
				// Proportionally divide the entry
				fracSending := sharesToSend.Quo(entry.SharesDst)
				balanceToSend := fracSending.MulInt(entry.InitialBalance).TruncateInt()
				balanceToKeep := entry.InitialBalance.Sub(balanceToSend)
				toRed, err := k.SetRedelegationEntry(
					ctx, toAddr, sdk.ValAddress(redelegation.ValidatorSrcAddress),
					sdk.ValAddress(redelegation.ValidatorDstAddress),
					entry.CreationHeight, entry.CompletionTime, balanceToSend, math.LegacyZeroDec(), sharesToSend,
				)
				if err != nil {
					redelegationErrors = append(redelegationErrors, err)
					continue
				}
				k.InsertRedelegationQueue(ctx, toRed, entry.CompletionTime)
				entry.InitialBalance = balanceToKeep
				entry.SharesDst = sharesToKeep
				redelegation.Entries[i] = entry
				// not modifying the completion time, so no need to change the queue
				redelegationModified = true
				entriesRemaining = true
			}
		}
		if redelegationModified {
			if entriesRemaining {
				err = k.SetRedelegation(ctx, redelegation)
				if err != nil {
					redelegationErrors = append(redelegationErrors, err)
					continue
				}
			} else {
				err = k.RemoveRedelegation(ctx, redelegation)
				if err != nil {
					redelegationErrors = append(redelegationErrors, err)
					continue
				}
			}
		}
	}

	if len(redelegationErrors) > 0 {
		return transferred, redelegationErrors[0]
	}
	return transferred, nil
}
