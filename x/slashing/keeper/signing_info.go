package keeper

import (
	"time"

	storetypes "cosmossdk.io/store/types"
	"github.com/bits-and-blooms/bitset"
	"github.com/cockroachdb/errors"

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

// getMissedBlockBitmapChunk gets the bitmap chunk at the given chunk index for
// a validator's missed block signing window.
func (k Keeper) getMissedBlockBitmapChunk(ctx sdk.Context, addr sdk.ConsAddress, chunkIndex int64) []byte {
	store := ctx.KVStore(k.storeKey)
	chunk := store.Get(types.ValidatorMissedBlockBitmapKey(addr, chunkIndex))
	return chunk
}

// setMissedBlockBitmapChunk sets the bitmap chunk at the given chunk index for
// a validator's missed block signing window.
func (k Keeper) setMissedBlockBitmapChunk(ctx sdk.Context, addr sdk.ConsAddress, chunkIndex int64, chunk []byte) {
	store := ctx.KVStore(k.storeKey)
	key := types.ValidatorMissedBlockBitmapKey(addr, chunkIndex)
	store.Set(key, chunk)
}

// GetMissedBlockBitmapValue returns true if a validator missed signing a block
// at the given index and false otherwise. The index provided is assumed to be
// the index in the range [0, SignedBlocksWindow), which represents the bitmap
// where each bit represents a height, and is determined by the validator's
// IndexOffset modulo SignedBlocksWindow. This index is used to fetch the chunk
// in the bitmap and the relative bit in that chunk.
func (k Keeper) GetMissedBlockBitmapValue(ctx sdk.Context, addr sdk.ConsAddress, index int64) (bool, error) {
	// get the chunk or "word" in the logical bitmap
	chunkIndex := index / types.MissedBlockBitmapChunkSize

	bs := bitset.New(uint(types.MissedBlockBitmapChunkSize))
	chunk := k.getMissedBlockBitmapChunk(ctx, addr, chunkIndex)
	if chunk != nil {
		if err := bs.UnmarshalBinary(chunk); err != nil {
			return false, errors.Wrapf(err, "failed to decode bitmap chunk; index: %d", index)
		}
	}

	// get the bit position in the chunk of the logical bitmap, where Test()
	// checks if the bit is set.
	bitIndex := index % types.MissedBlockBitmapChunkSize
	return bs.Test(uint(bitIndex)), nil
}

// SetMissedBlockBitmapValue sets, i.e. flips, a bit in the validator's missed
// block bitmap. When missed=true, the bit is set, otherwise it set to zero. The
// index provided is assumed to be the index in the range [0, SignedBlocksWindow),
// which represents the bitmap where each bit represents a height, and is
// determined by the validator's IndexOffset modulo SignedBlocksWindow. This
// index is used to fetch the chunk in the bitmap and the relative bit in that
// chunk.
func (k Keeper) SetMissedBlockBitmapValue(ctx sdk.Context, addr sdk.ConsAddress, index int64, missed bool) error {
	// get the chunk or "word" in the logical bitmap
	chunkIndex := index / types.MissedBlockBitmapChunkSize

	bs := bitset.New(uint(types.MissedBlockBitmapChunkSize))
	chunk := k.getMissedBlockBitmapChunk(ctx, addr, chunkIndex)
	if chunk != nil {
		if err := bs.UnmarshalBinary(chunk); err != nil {
			return errors.Wrapf(err, "failed to decode bitmap chunk; index: %d", index)
		}
	}

	// get the bit position in the chunk of the logical bitmap
	bitIndex := uint(index % types.MissedBlockBitmapChunkSize)
	if missed {
		bs.Set(bitIndex)
	} else {
		bs.Clear(bitIndex)
	}

	updatedChunk, err := bs.MarshalBinary()
	if err != nil {
		return errors.Wrapf(err, "failed to encode bitmap chunk; index: %d", index)
	}

	k.setMissedBlockBitmapChunk(ctx, addr, chunkIndex, updatedChunk)
	return nil
}

// DeleteMissedBlockBitmap removes a validator's missed block bitmap from state.
func (k Keeper) DeleteMissedBlockBitmap(ctx sdk.Context, addr sdk.ConsAddress) {
	store := ctx.KVStore(k.storeKey)

	iter := storetypes.KVStorePrefixIterator(store, types.ValidatorMissedBlockBitmapPrefixKey(addr))
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		store.Delete(iter.Key())
	}
}

// IterateMissedBlockBitmap iterates over a validator's signed blocks window
// bitmap and performs a callback function on each index, i.e. block height, in
// the range [0, SignedBlocksWindow).
//
// Note: A callback will only be executed over all bitmap chunks that exist in
// state.
func (k Keeper) IterateMissedBlockBitmap(ctx sdk.Context, addr sdk.ConsAddress, cb func(index int64, missed bool) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iter := storetypes.KVStorePrefixIterator(store, types.ValidatorMissedBlockBitmapPrefixKey(addr))
	defer iter.Close()

	var index int64
	for ; iter.Valid(); iter.Next() {
		bs := bitset.New(uint(types.MissedBlockBitmapChunkSize))

		if err := bs.UnmarshalBinary(iter.Value()); err != nil {
			panic(errors.Wrapf(err, "failed to decode bitmap chunk; index: %v", string(iter.Key())))
		}

		for i := uint(0); i < types.MissedBlockBitmapChunkSize; i++ {
			// execute the callback, where Test() returns true if the bit is set
			if cb(index, bs.Test(i)) {
				break
			}

			index++
		}
	}
}

// GetValidatorMissedBlocks returns array of missed blocks for given validator.
func (k Keeper) GetValidatorMissedBlocks(ctx sdk.Context, addr sdk.ConsAddress) []types.MissedBlock {
	missedBlocks := make([]types.MissedBlock, 0, k.SignedBlocksWindow(ctx))
	k.IterateMissedBlockBitmap(ctx, addr, func(index int64, missed bool) (stop bool) {
		if missed {
			missedBlocks = append(missedBlocks, types.NewMissedBlock(index, missed))
		}

		return false
	})

	return missedBlocks
}
