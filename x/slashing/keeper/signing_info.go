package keeper

import (
	"time"

	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
)

// GetValidatorSigningInfo retruns the ValidatorSigningInfo for a specific validator
// ConsAddress
func (k Keeper) GetValidatorSigningInfo(ctx sdk.Context, address sdk.ConsAddress) (types.ValidatorSigningInfo, bool) {
	store := ctx.KVStore(k.storeKey)

	var info types.ValidatorSigningInfo
	bz := store.Get(types.ValidatorSigningInfoKey(address))
	if bz == nil {
		return info, false
	}

	k.cdc.MustUnmarshal(bz, &info)
	return info, true
}

// HasValidatorSigningInfo returns if a given validator has signing information
// persisted.
func (k Keeper) HasValidatorSigningInfo(ctx sdk.Context, consAddr sdk.ConsAddress) bool {
	_, ok := k.GetValidatorSigningInfo(ctx, consAddr)
	return ok
}

// SetValidatorSigningInfo sets the validator signing info to a consensus address key
func (k Keeper) SetValidatorSigningInfo(ctx sdk.Context, address sdk.ConsAddress, info types.ValidatorSigningInfo) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&info)
	store.Set(types.ValidatorSigningInfoKey(address), bz)
}

// IterateValidatorSigningInfos iterates over the stored ValidatorSigningInfo
func (k Keeper) IterateValidatorSigningInfos(ctx sdk.Context,
	handler func(address sdk.ConsAddress, info types.ValidatorSigningInfo) (stop bool),
) {
	store := ctx.KVStore(k.storeKey)
	iter := storetypes.KVStorePrefixIterator(store, types.ValidatorSigningInfoKeyPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		address := types.ValidatorSigningInfoAddress(iter.Key())
		var info types.ValidatorSigningInfo
		k.cdc.MustUnmarshal(iter.Value(), &info)
		if handler(address, info) {
			break
		}
	}
}

// // GetValidatorMissedBlockBitArray gets the bit for the missed blocks array
// func (k Keeper) GetValidatorMissedBlockBitArray(ctx sdk.Context, address sdk.ConsAddress, index int64) bool {
// 	store := ctx.KVStore(k.storeKey)
// 	bz := store.Get(types.ValidatorMissedBlockBitArrayKey(address, index))
// 	var missed gogotypes.BoolValue
// 	if bz == nil {
// 		// lazy: treat empty key as not missed
// 		return false
// 	}
// 	k.cdc.MustUnmarshal(bz, &missed)

// 	return missed.Value
// }

// // IterateValidatorMissedBlockBitArray iterates over the signed blocks window
// // and performs a callback function
// func (k Keeper) IterateValidatorMissedBlockBitArray(ctx sdk.Context,
// 	address sdk.ConsAddress, handler func(index int64, missed bool) (stop bool),
// ) {
// 	store := ctx.KVStore(k.storeKey)
// 	index := int64(0)
// 	// Array may be sparse
// 	for ; index < k.SignedBlocksWindow(ctx); index++ {
// 		var missed gogotypes.BoolValue
// 		bz := store.Get(types.ValidatorMissedBlockBitArrayKey(address, index))
// 		if bz == nil {
// 			continue
// 		}

// 		k.cdc.MustUnmarshal(bz, &missed)
// 		if handler(index, missed.Value) {
// 			break
// 		}
// 	}
// }

// // GetValidatorMissedBlocks returns array of missed blocks for given validator Cons address
// func (k Keeper) GetValidatorMissedBlocks(ctx sdk.Context, address sdk.ConsAddress) []types.MissedBlock {
// 	missedBlocks := []types.MissedBlock{}
// 	k.IterateValidatorMissedBlockBitArray(ctx, address, func(index int64, missed bool) (stop bool) {
// 		missedBlocks = append(missedBlocks, types.NewMissedBlock(index, missed))
// 		return false
// 	})

// 	return missedBlocks
// }

// // SetValidatorMissedBlockBitArray sets the bit that checks if the validator has
// // missed a block in the current window
// func (k Keeper) SetValidatorMissedBlockBitArray(ctx sdk.Context, address sdk.ConsAddress, index int64, missed bool) {
// 	store := ctx.KVStore(k.storeKey)
// 	bz := k.cdc.MustMarshal(&gogotypes.BoolValue{Value: missed})
// 	store.Set(types.ValidatorMissedBlockBitArrayKey(address, index), bz)
// }

// // clearValidatorMissedBlockBitArray deletes every instance of ValidatorMissedBlockBitArray in the store
// func (k Keeper) clearValidatorMissedBlockBitArray(ctx sdk.Context, address sdk.ConsAddress) {
// 	store := ctx.KVStore(k.storeKey)
// 	iter := storetypes.KVStorePrefixIterator(store, types.ValidatorMissedBlockBitArrayPrefixKey(address))
// 	defer iter.Close()
// 	for ; iter.Valid(); iter.Next() {
// 		store.Delete(iter.Key())
// 	}
// }

// JailUntil attempts to set a validator's JailedUntil attribute in its signing
// info. It will panic if the signing info does not exist for the validator.
func (k Keeper) JailUntil(ctx sdk.Context, consAddr sdk.ConsAddress, jailTime time.Time) {
	signInfo, ok := k.GetValidatorSigningInfo(ctx, consAddr)
	if !ok {
		panic("cannot jail validator that does not have any signing information")
	}

	signInfo.JailedUntil = jailTime
	k.SetValidatorSigningInfo(ctx, consAddr, signInfo)
}

// Tombstone attempts to tombstone a validator. It will panic if signing info for
// the given validator does not exist.
func (k Keeper) Tombstone(ctx sdk.Context, consAddr sdk.ConsAddress) {
	signInfo, ok := k.GetValidatorSigningInfo(ctx, consAddr)
	if !ok {
		panic("cannot tombstone validator that does not have any signing information")
	}

	if signInfo.Tombstoned {
		panic("cannot tombstone validator that is already tombstoned")
	}

	signInfo.Tombstoned = true
	k.SetValidatorSigningInfo(ctx, consAddr, signInfo)
}

// IsTombstoned returns if a given validator by consensus address is tombstoned.
func (k Keeper) IsTombstoned(ctx sdk.Context, consAddr sdk.ConsAddress) bool {
	signInfo, ok := k.GetValidatorSigningInfo(ctx, consAddr)
	if !ok {
		return false
	}

	return signInfo.Tombstoned
}

// ============================================================================

// GetMissedBlockBitmapValue returns a validator's missed block bitmap value at
// the given index, where index is the block height offset in the signing window.
func (k Keeper) GetMissedBlockBitmapValue(ctx sdk.Context, addr sdk.ConsAddress, index int64) bool {
	// assume index provided is non-zero based
	pos := index - 1

	// get the chunk or "word" in the logical bitmap
	chunkIndex := pos / types.MissedBlockBitmapChunkSize

	// get the position in the chunk of the logical bitmap
	bitIndex := pos % types.MissedBlockBitmapChunkSize

	store := ctx.KVStore(k.storeKey)
	chunk := store.Get(types.ValidatorMissedBlockBitmapKey(addr, chunkIndex))

	return chunk[bitIndex] == 1
}

// SetMissedBlockBitmapValue sets a validator's missed block bitmap value at the
// given index, where index is the block height offset in the signing window.
// If missed=true, the bit is set to 1, otherwise it is set to 0.
func (k Keeper) SetMissedBlockBitmapValue(ctx sdk.Context, addr sdk.ConsAddress, index int64, missed bool) {
	// assume index provided is non-zero based
	pos := index - 1

	// get the chunk or "word" in the logical bitmap
	chunkIndex := pos / types.MissedBlockBitmapChunkSize

	// get the position in the chunk of the logical bitmap
	bitIndex := pos % types.MissedBlockBitmapChunkSize

	store := ctx.KVStore(k.storeKey)
	key := types.ValidatorMissedBlockBitmapKey(addr, chunkIndex)

	chunk := store.Get(key)
	if missed {
		chunk[bitIndex] = 1
	} else {
		chunk[bitIndex] = 0
	}

	store.Set(key, chunk)
}
