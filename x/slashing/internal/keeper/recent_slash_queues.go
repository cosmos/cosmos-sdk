package keeper

import (
	"time"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/internal/types"
)

// RecentSlashesQueues

// Get the prefix store for the Recent Double Signs Queue
func (k Keeper) DoubleSignQueueStore(ctx sdk.Context) sdk.KVStore {
	return prefix.NewStore(ctx.KVStore(k.storeKey), []byte("dsqueue"))
}

// InsertDoubleSignQueue inserts a double sign event into the queue at time of doublesign
func (k Keeper) InsertDoubleSignQueue(ctx sdk.Context, slashEvent types.SlashEvent, removeTime time.Time) {
	dsStore := k.DoubleSignQueueStore(ctx)
	bz := k.cdc.MustMarshalBinaryBare(slashEvent)
	dsStore.Set(sdk.FormatTimeBytes(removeTime), bz)
}

// Get the prefix store for the Recent Liveness Faults Queue
func (k Keeper) LivenessQueueStore(ctx sdk.Context) sdk.KVStore {
	return prefix.NewStore(ctx.KVStore(k.storeKey), []byte("livequeue"))
}

// InsertDoubleSignQueue inserts a double sign event into the queue at time of doublesign
func (k Keeper) InsertLivenessQueue(ctx sdk.Context, slashEvent types.SlashEvent, removeTime time.Time) {
	liveStore := k.LivenessQueueStore(ctx)
	bz := k.cdc.MustMarshalBinaryBare(slashEvent)
	liveStore.Set(sdk.FormatTimeBytes(removeTime), bz)
}

// Iterators

// IterateDoubleSignQueue iterates over the slash events in the recent double signs queue
// and performs a callback function
func (k Keeper) IterateDoubleSignQueue(ctx sdk.Context, cb func(slashEvent types.SlashEvent) (stop bool)) {
	iterator := k.DoubleSignQueueIterator(ctx)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var slashEvent types.SlashEvent
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &slashEvent)

		if cb(slashEvent) {
			break
		}
	}
}

// IterateLivenessQueue iterates over the slash events in the recent liveness faults queue
// and performs a callback function
func (k Keeper) IterateLivenessQueue(ctx sdk.Context, cb func(slashEvent types.SlashEvent) (stop bool)) {
	iterator := k.LivenessQueueIterator(ctx)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var slashEvent types.SlashEvent
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &slashEvent)

		if cb(slashEvent) {
			break
		}
	}
}

// ActiveProposalQueueIterator returns an sdk.Iterator for all the proposals in the Active Queue that expire by endTime
func (k Keeper) DoubleSignQueueIterator(ctx sdk.Context) sdk.Iterator {
	dsStore := k.DoubleSignQueueStore(ctx)
	return dsStore.Iterator(nil, nil)
}

// InactiveProposalQueueIterator returns an sdk.Iterator for all the proposals in the Inactive Queue that expire by endTime
func (k Keeper) LivenessQueueIterator(ctx sdk.Context) sdk.Iterator {
	liveStore := k.LivenessQueueStore(ctx)
	return liveStore.Iterator(nil, nil)
}
