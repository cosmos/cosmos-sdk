package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	proto "github.com/gogo/protobuf/proto"
)

// keys
var (
	NextEpochActionIDPrefix = []byte("next_epoch_action_id")
	EpochActionStorePrefix  = "epoch_action"
)

// SetNextEpochActionID save ID to be used for next epoch action
func (k Keeper) SetNextEpochActionID(ctx sdk.Context, actionID uint64) {
	store := ctx.KVStore(k.storeKey)

	store.Set(NextEpochActionIDPrefix, sdk.Uint64ToBigEndian(actionID))
}

// GetNextEpochActionID returns ID to be used for next epoch
func (k Keeper) GetNextEpochActionID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(NextEpochActionIDPrefix)
	if bz == nil {
		// return default action ID to 1
		return 1
	}

	return sdk.BigEndianToUint64(bz)
}

// ActionStoreKey returns action store key from ID
func ActionStoreKey(actionID uint64) []byte {
	return []byte(fmt.Sprintf("%s_%d", EpochActionStorePrefix, actionID))
}

// SaveEpochAction save the actions that need to be executed on next epoch
func (k Keeper) SaveEpochAction(ctx sdk.Context, action proto.Message) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshalBinaryBare(action)
	actionID := k.GetNextEpochActionID(ctx)
	store.Set(ActionStoreKey(actionID), bz)
	k.SetNextEpochActionID(ctx, actionID+1)
}

// GetEpochAction get action by ID
func (k Keeper) GetEpochAction(ctx sdk.Context, actionID uint64) proto.Message {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(ActionStoreKey(actionID))
	if bz == nil {
		return nil
	}

	var action proto.Message
	k.cdc.MustUnmarshalBinaryBare(bz, &action)

	return action
}

// GetEpochActions get all actions
func (k Keeper) GetEpochActions(ctx sdk.Context) []proto.Message {
	actions := []proto.Message{}
	iterator := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), []byte(EpochActionStorePrefix))

	for ; iterator.Valid(); iterator.Next() {
		var action proto.Message
		// TODO is this correct to use proto.Message for serialization?
		bz := iterator.Value()
		k.cdc.MustUnmarshalBinaryBare(bz, &action)
		actions = append(actions, action)
	}

	return actions
}

// DequeueEpochActions dequeue all the actions store on epoch
func (k Keeper) DequeueEpochActions(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, []byte(EpochActionStorePrefix))

	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		store.Delete(key)
	}
}
