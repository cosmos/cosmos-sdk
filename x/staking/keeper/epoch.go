package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	db "github.com/tendermint/tm-db"
)

// keys
var (
	NextEpochActionID      = []byte("next_epoch_action_id")
	EpochActionQueuePrefix = "epoch_action"
)

// SetNextEpochActionID save ID to be used for next epoch action
func (k Keeper) SetNextEpochActionID(ctx sdk.Context, actionID uint64) {
	store := ctx.KVStore(k.storeKey)

	store.Set(NextEpochActionID, sdk.Uint64ToBigEndian(actionID))
}

// GetNextEpochActionID returns ID to be used for next epoch
func (k Keeper) GetNextEpochActionID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(NextEpochActionID)
	if bz == nil {
		// return default action ID to 1
		return 1
	}

	return sdk.BigEndianToUint64(bz)
}

// ActionStoreKey returns action store key from ID
func ActionStoreKey(epochIndex uint64, actionID uint64) []byte {
	return []byte(fmt.Sprintf("%s_%d_%d", EpochActionQueuePrefix, epochIndex, actionID))
}

// SaveEpochAction save the actions that need to be executed on next epoch
func (k Keeper) SaveEpochAction(ctx sdk.Context, epochIndex uint64, action sdk.Msg) {
	store := ctx.KVStore(k.storeKey)

	// reference from TestMarshalAny(t *testing.T)
	bz, err := codec.MarshalAny(k.cdc, action)
	if err != nil {
		panic(err)
	}
	actionID := k.GetNextEpochActionID(ctx)
	store.Set(ActionStoreKey(epochIndex, actionID), bz)
	k.SetNextEpochActionID(ctx, actionID+1)
}

// GetEpochAction get action by ID
func (k Keeper) GetEpochAction(ctx sdk.Context, epochIndex uint64, actionID uint64) sdk.Msg {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(ActionStoreKey(epochIndex, actionID))
	if bz == nil {
		return nil
	}

	var action sdk.Msg
	// reference from TestMarshalAny(t *testing.T)
	codec.UnmarshalAny(k.cdc, &action, bz)

	return action
}

// GetEpochActions get all actions
func (k Keeper) GetEpochActions(ctx sdk.Context) []sdk.Msg {
	actions := []sdk.Msg{}
	iterator := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), []byte(EpochActionQueuePrefix))

	for ; iterator.Valid(); iterator.Next() {
		var action sdk.Msg
		bz := iterator.Value()
		// reference from TestMarshalAny(t *testing.T)
		codec.UnmarshalAny(k.cdc, &action, bz)
		actions = append(actions, action)
	}

	return actions
}

// GetEpochActionsIteratorByEpochIndex returns iterator for EpochActions
func (k Keeper) GetEpochActionsIteratorByEpochIndex(ctx sdk.Context, epochIndex uint64) db.Iterator {
	prefixKey := fmt.Sprintf("%s_%d", EpochActionQueuePrefix, epochIndex)
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), []byte(prefixKey))
}

// DequeueEpochActions dequeue all the actions store on epoch
func (k Keeper) DequeueEpochActions(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, []byte(EpochActionQueuePrefix))

	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		store.Delete(key)
	}
}

// DeleteByKey delete item by key
func (k Keeper) DeleteByKey(ctx sdk.Context, key []byte) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(key)
}

// GetEpochActionByIterator get action by iterator
func (k Keeper) GetEpochActionByIterator(iterator db.Iterator) sdk.Msg {
	bz := iterator.Value()

	var action sdk.Msg
	// reference from TestMarshalAny(t *testing.T)
	codec.UnmarshalAny(k.cdc, &action, bz)

	return action
}
