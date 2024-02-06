package keeper

import (
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	db "github.com/tendermint/tm-db"
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

// GetNewActionID returns ID to be used for next epoch
func (k Keeper) GetNewActionID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(NextEpochActionID)
	if bz == nil {
		// return default action ID to 1
		return DefaultEpochActionID
	}
	id := sdk.BigEndianToUint64(bz)

	// increment next action ID
	store.Set(NextEpochActionID, sdk.Uint64ToBigEndian(id+1))

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
	store.Set(ActionStoreKey(epochNumber, actionID), bz)
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
	store.Set(ActionStoreKey(epochNumber, actionID), bz)
}

// GetEpochMsg gets a msg by ID
func (k Keeper) GetEpochMsg(ctx sdk.Context, epochNumber int64, actionID uint64) sdk.Msg {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(ActionStoreKey(epochNumber, actionID))
	if bz == nil {
		return nil
	}

	var action sdk.Msg
	k.cdc.UnmarshalInterface(bz, &action)

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
	k.cdc.UnmarshalInterface(bz, &action)

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
		return DefaultEpochNumber
	}

	return int64(sdk.BigEndianToUint64(bz))
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
