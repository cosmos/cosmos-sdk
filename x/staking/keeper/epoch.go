package keeper

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	db "github.com/tendermint/tm-db"
)

// keys
var (
	NextEpochActionID      = []byte("next_epoch_action_id")
	EpochNumberID          = []byte("epoch_number_id")
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
func ActionStoreKey(epochNumber int64, actionID uint64) []byte {
	return []byte(fmt.Sprintf("%s_%d_%d", EpochActionQueuePrefix, epochNumber, actionID))
}

// QueueMsgForEpoch save the actions that need to be executed on next epoch
func (k Keeper) QueueMsgForEpoch(ctx sdk.Context, epochNumber int64, action sdk.Msg) {
	store := ctx.KVStore(k.storeKey)

	// reference from TestMarshalAny(t *testing.T)
	bz, err := codec.MarshalAny(k.cdc, action)
	if err != nil {
		panic(err)
	}
	actionID := k.GetNextEpochActionID(ctx)
	store.Set(ActionStoreKey(epochNumber, actionID), bz)
	k.SetNextEpochActionID(ctx, actionID+1)
}

// RestoreEpochAction restore the actions that need to be exectued on next epoch
func (k Keeper) RestoreEpochAction(ctx sdk.Context, epochNumber int64, action *codectypes.Any) {
	store := ctx.KVStore(k.storeKey)

	// reference from TestMarshalAny(t *testing.T)
	bz, err := codec.MarshalAny(k.cdc, action)
	if err != nil {
		panic(err)
	}
	actionID := k.GetNextEpochActionID(ctx)
	store.Set(ActionStoreKey(epochNumber, actionID), bz)
	k.SetNextEpochActionID(ctx, actionID+1)
}

// GetEpochAction get action by ID
func (k Keeper) GetEpochAction(ctx sdk.Context, epochNumber int64, actionID uint64) sdk.Msg {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(ActionStoreKey(epochNumber, actionID))
	if bz == nil {
		return nil
	}

	var action sdk.Msg
	// reference from TestMarshalAny(t *testing.T)
	codec.UnmarshalAny(k.cdc, &action, bz)

	return action
}

// GetEpochActions get all actions
func (k Keeper) GetEpochActions(ctx sdk.Context) []*codectypes.Any {
	actions := []*codectypes.Any{}
	iterator := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), []byte(EpochActionQueuePrefix))

	for ; iterator.Valid(); iterator.Next() {
		var action codectypes.Any
		bz := iterator.Value()
		// reference from TestMarshalAny(t *testing.T)
		codec.UnmarshalAny(k.cdc, &action, bz)
		actions = append(actions, &action)
	}

	return actions
}

// GetEpochActionsIterator returns iterator for EpochActions
func (k Keeper) GetEpochActionsIterator(ctx sdk.Context) db.Iterator {
	prefixKey := fmt.Sprintf("%s", EpochActionQueuePrefix)
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

// SetEpochNumber set epoch number
func (k Keeper) SetEpochNumber(ctx sdk.Context, epochNumber int64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(EpochNumberID, sdk.Uint64ToBigEndian(uint64(epochNumber)))
}

// GetEpochNumber fetches epoch number
func (k Keeper) GetEpochNumber(ctx sdk.Context) int64 {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(EpochNumberID)
	if bz == nil {
		// return default EpochNumber 0
		return 0
	}

	return int64(sdk.BigEndianToUint64(bz))
}

// IncreaseEpochNumber increases epoch number
func (k Keeper) IncreaseEpochNumber(ctx sdk.Context) {
	epochNumber := k.GetEpochNumber(ctx)
	k.SetEpochNumber(ctx, epochNumber+1)
}

// GetNextEpochHeight returns next epoch block height
func (k Keeper) GetNextEpochHeight(ctx sdk.Context) int64 {
	currentHeight := ctx.BlockHeight()
	epochInterval := k.EpochInterval(ctx)
	return currentHeight + (epochInterval - currentHeight%epochInterval)
}

// GetNextEpochTime returns estimated next epoch time
func (k Keeper) GetNextEpochTime(ctx sdk.Context) time.Time {
	currentTime := ctx.BlockTime()
	currentHeight := ctx.BlockHeight()
	timeoutCommit := 5 * time.Second // TODO how to get timeout commit tendermint config?
	// cp := baseapp.GetConsensusParams(ctx)

	return currentTime.Add(timeoutCommit * time.Duration(k.GetNextEpochHeight(ctx)-currentHeight))
}
