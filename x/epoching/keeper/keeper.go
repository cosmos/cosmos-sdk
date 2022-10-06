package keeper

import (
	"time"

	db "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	store2 "github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	DefaultEpochActionID = 1
	DefaultEpochNumber   = 0
)

var (
	NextEpochActionID      = []byte{0x11}
	EpochNumberID          = []byte{0x12}
	EpochActionQueuePrefix = []byte{0x13} // prefix for the epoch
)

// Keeper of the store
type Keeper struct {
	storeKey storetypes.StoreKey
	cdc      codec.BinaryCodec
	// Used to calculate the estimated next epoch time.
	// This is local to every node
	// TODO: remove in favor of consensus param when its added
	commitTimeout time.Duration
}

// NewKeeper creates a epoch queue manager
func NewKeeper(cdc codec.BinaryCodec, key storetypes.StoreKey, commitTimeout time.Duration) Keeper {
	return Keeper{
		storeKey:      key,
		cdc:           cdc,
		commitTimeout: commitTimeout,
	}
}

func (k Keeper) decodeID(bz []byte) (uint64, error) {
	if bz == nil {
		// return default action ID to 1
		return DefaultEpochActionID, nil
	}
	id := sdk.BigEndianToUint64(bz)
	return id, nil
}

// GetNewActionID returns ID to be used for next epoch
func (k Keeper) GetNewActionID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	id, err := store2.GetAndDecode(store, k.decodeID, NextEpochActionID)
	if err != nil {
		panic(err)
	}

	// increment next action ID
	store2.Set(store, NextEpochActionID, sdk.Uint64ToBigEndian(id+1))

	return id
}

// ActionStoreKey returns action store key from ID
func ActionStoreKey(epochNumber int64, actionID uint64) []byte {
	return append(EpochActionQueuePrefix, byte(epochNumber), byte(actionID))
}

// QueueMsgForEpoch save the actions that need to be executed on next epoch
func (k Keeper) QueueMsgForEpoch(ctx sdk.Context, epochNumber int64, msg sdk.Msg) {
	store := ctx.KVStore(k.storeKey)

	bz, err := k.cdc.MarshalInterface(msg)
	if err != nil {
		panic(err)
	}

	actionID := k.GetNewActionID(ctx)
	store2.Set(store, ActionStoreKey(epochNumber, actionID), bz)
}

// RestoreEpochAction restore the actions that need to be executed on next epoch
func (k Keeper) RestoreEpochAction(ctx sdk.Context, epochNumber int64, action *codectypes.Any) {
	store := ctx.KVStore(k.storeKey)

	// reference from TestMarshalAny(t *testing.T)
	bz, err := k.cdc.MarshalInterface(action)
	if err != nil {
		panic(err)
	}

	actionID := k.GetNewActionID(ctx)
	store2.Set(store, ActionStoreKey(epochNumber, actionID), bz)
}

func (k Keeper) decodeMsg(bz []byte) (sdk.Msg, error) {
	if bz == nil {
		return nil, nil
	}
	var action sdk.Msg
	err := k.cdc.UnmarshalInterface(bz, &action)
	if err != nil {
		panic(err)
	}
	return action, nil
}

// GetEpochMsg gets a msg by ID
func (k Keeper) GetEpochMsg(ctx sdk.Context, epochNumber int64, actionID uint64) sdk.Msg {
	store := ctx.KVStore(k.storeKey)

	action, err := store2.GetAndDecode(store, k.decodeMsg, ActionStoreKey(epochNumber, actionID))
	if err != nil {
		return nil
	}

	return action
}

// GetEpochActions get all actions
func (k Keeper) GetEpochActions(ctx sdk.Context) []sdk.Msg {
	actions := []sdk.Msg{}
	iterator := k.GetEpochActionsIterator(ctx)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var action sdk.Msg
		bz := iterator.Value()
		k.cdc.UnmarshalInterface(bz, &action)
		actions = append(actions, action)
	}

	return actions
}

// GetEpochActionsIterator returns iterator for EpochActions
func (k Keeper) GetEpochActionsIterator(ctx sdk.Context) db.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), EpochActionQueuePrefix)
}

// DequeueEpochActions dequeue all the actions store on epoch
func (k Keeper) DequeueEpochActions(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, EpochActionQueuePrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		store2.Delete(store, key)
	}
}

// DeleteByKey delete item by key
func (k Keeper) DeleteByKey(ctx sdk.Context, key []byte) {
	store := ctx.KVStore(k.storeKey)
	store2.Delete(store, key)
}

// GetEpochActionByIterator get action by iterator
func (k Keeper) GetEpochActionByIterator(iterator db.Iterator) sdk.Msg {
	bz := iterator.Value()

	var action sdk.Msg
	k.cdc.UnmarshalInterface(bz, &action)

	return action
}

// SetEpochNumber set epoch number
func (k Keeper) SetEpochNumber(ctx sdk.Context, epochNumber int64) {
	store := ctx.KVStore(k.storeKey)
	store2.Set(store, EpochNumberID, sdk.Uint64ToBigEndian(uint64(epochNumber)))
}

func (k Keeper) decodeNumberID(bz []byte) (int64, error) {
	if bz == nil {
		return DefaultEpochNumber, nil
	}
	return int64(sdk.BigEndianToUint64(bz)), nil
}

// GetEpochNumber fetches epoch number
func (k Keeper) GetEpochNumber(ctx sdk.Context) int64 {
	store := ctx.KVStore(k.storeKey)

	epochNumberID, err := store2.GetAndDecode(store, k.decodeNumberID, EpochNumberID)
	if err != nil {
		panic(err)
	}
	return epochNumberID
}

// IncreaseEpochNumber increases epoch number
func (k Keeper) IncreaseEpochNumber(ctx sdk.Context) {
	epochNumber := k.GetEpochNumber(ctx)
	k.SetEpochNumber(ctx, epochNumber+1)
}

// GetNextEpochHeight returns next epoch block height
func (k Keeper) GetNextEpochHeight(ctx sdk.Context, epochInterval int64) int64 {
	currentHeight := ctx.BlockHeight()
	return currentHeight + (epochInterval - currentHeight%epochInterval)
}

// GetNextEpochTime returns estimated next epoch time
func (k Keeper) GetNextEpochTime(ctx sdk.Context, epochInterval int64) time.Time {
	currentTime := ctx.BlockTime()
	currentHeight := ctx.BlockHeight()

	return currentTime.Add(k.commitTimeout * time.Duration(k.GetNextEpochHeight(ctx, epochInterval)-currentHeight))
}
