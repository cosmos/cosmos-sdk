package keeper

import (
	"time"

	db "github.com/tendermint/tm-db"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetNewActionID returns ID to be used for next message queue item
func (k Keeper) GetNewActionID(ctx sdk.Context) uint64 {
	return k.epochKeeper.GetNewActionID(ctx)
}

// QueueMsgForEpoch save the actions that need to be executed on next epoch
func (k Keeper) QueueMsgForEpoch(ctx sdk.Context, epochNumber int64, action sdk.Msg) {
	k.epochKeeper.QueueMsgForEpoch(ctx, epochNumber, action)
}

// RestoreEpochMsg restore the actions that need to be executed on next epoch
func (k Keeper) RestoreEpochMsg(ctx sdk.Context, epochNumber int64, action *codectypes.Any) {
	k.epochKeeper.RestoreEpochMsg(ctx, epochNumber, action)
}

// GetEpochMsg get action by ID
func (k Keeper) GetEpochMsg(ctx sdk.Context, epochNumber int64, msgID uint64) sdk.Msg {
	return k.epochKeeper.GetEpochMsg(ctx, epochNumber, msgID)
}

// GetEpochMsgs get all actions
func (k Keeper) GetEpochMsgs(ctx sdk.Context) []sdk.Msg {
	return k.epochKeeper.GetEpochMsgs(ctx)
}

// GetEpochMsgByIterator get action by iterator
func (k Keeper) GetEpochMsgByIterator(iterator db.Iterator) sdk.Msg {
	return k.epochKeeper.GetEpochMsgByIterator(iterator)
}

// SetEpochNumber set epoch number
func (k Keeper) SetEpochNumber(ctx sdk.Context, epochNumber int64) {
	k.epochKeeper.SetEpochNumber(ctx, epochNumber)
}

// GetEpochNumber fetches epoch number
func (k Keeper) GetEpochNumber(ctx sdk.Context) int64 {
	return k.epochKeeper.GetEpochNumber(ctx)
}

// GetNextEpochHeight returns next epoch block height
func (k Keeper) GetNextEpochHeight(ctx sdk.Context) int64 {
	return k.epochKeeper.GetNextEpochHeight(ctx, k.EpochInterval(ctx))
}

// GetNextEpochTime returns estimated next epoch time
func (k Keeper) GetNextEpochTime(ctx sdk.Context) time.Time {
	return k.epochKeeper.GetNextEpochTime(ctx, k.EpochInterval(ctx))
}
