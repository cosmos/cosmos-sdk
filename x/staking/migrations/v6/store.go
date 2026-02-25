package v6

import (
	"context"
	"fmt"
	"time"

	storetypes "cosmossdk.io/core/store"
	storetypesv1 "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// QueuePendingSlotSetter defines the interface for setting queue pending slots.
// This interface is implemented by keeper.Keeper to avoid import cycles.
type QueuePendingSlotSetter interface {
	SetValidatorQueuePendingSlots(ctx context.Context, slots []types.TimeHeightQueueSlot) error
	SetUBDQueuePendingSlots(ctx context.Context, slots []time.Time) error
	SetRedelegationQueuePendingSlots(ctx context.Context, slots []time.Time) error
}

// MigrateStore performs in-place store migrations from v5 to v6 by
// populating the queue pending-slot indexes (validator, UBD, redelegation)
// from current queue state. This avoids expensive full-range iteration in
// end-block on the first block after upgrade.
func MigrateStore(
	ctx sdk.Context,
	store storetypes.KVStore,
	k QueuePendingSlotSetter,
) error {
	if err := PopulateValidatorQueuePendingFromIterator(ctx, store, k.SetValidatorQueuePendingSlots); err != nil {
		return err
	}
	if err := PopulateUBDQueuePendingFromIterator(ctx, store, k.SetUBDQueuePendingSlots); err != nil {
		return err
	}
	return PopulateRedelegationQueuePendingFromIterator(ctx, store, k.SetRedelegationQueuePendingSlots)
}

func PopulateValidatorQueuePendingFromIterator(
	ctx context.Context,
	store storetypes.KVStore,
	setter func(context.Context, []types.TimeHeightQueueSlot) error,
) error {
	iter, err := store.Iterator(types.ValidatorQueueKey, storetypesv1.PrefixEndBytes(types.ValidatorQueueKey))
	if err != nil {
		return err
	}
	defer iter.Close()
	var slots []types.TimeHeightQueueSlot
	for ; iter.Valid(); iter.Next() {
		keyTime, keyHeight, err := types.ParseValidatorQueueKey(iter.Key())
		if err != nil {
			return err
		}
		slots = append(slots, types.TimeHeightQueueSlot{Time: keyTime, Height: keyHeight})
	}
	return setter(ctx, slots)
}

func populateTimeQueuePendingFromIterator(
	ctx context.Context,
	store storetypes.KVStore,
	queueKey []byte,
	setter func(context.Context, []time.Time) error,
) error {
	iter, err := store.Iterator(queueKey, storetypesv1.PrefixEndBytes(queueKey))
	if err != nil {
		return err
	}
	defer iter.Close()
	var slots []time.Time
	for ; iter.Valid(); iter.Next() {
		key := iter.Key()
		if len(key) <= len(queueKey) {
			return fmt.Errorf("key length is too short")
		}
		timeBz := key[len(queueKey):]
		t, err := sdk.ParseTimeBytes(timeBz)
		if err != nil {
			return fmt.Errorf("unable to parse time from queue key %x: %w", key, err)
		}
		slots = append(slots, t)
	}
	return setter(ctx, slots)
}

func PopulateUBDQueuePendingFromIterator(
	ctx context.Context,
	store storetypes.KVStore,
	setter func(context.Context, []time.Time) error,
) error {
	return populateTimeQueuePendingFromIterator(ctx, store, types.UnbondingQueueKey, setter)
}

func PopulateRedelegationQueuePendingFromIterator(
	ctx context.Context,
	store storetypes.KVStore,
	setter func(context.Context, []time.Time) error,
) error {
	return populateTimeQueuePendingFromIterator(ctx, store, types.RedelegationQueueKey, setter)
}
