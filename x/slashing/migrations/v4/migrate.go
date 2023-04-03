package v4

import (
	"cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"
	"github.com/bits-and-blooms/bitset"
	gogotypes "github.com/cosmos/gogoproto/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
)

// Migrate migrates state to consensus version 4. Specifically, the migration
// deletes all existing validator bitmap entries and replaces them with a real
// "chunked" bitmap.
func Migrate(ctx sdk.Context, cdc codec.BinaryCodec, store storetypes.KVStore, params types.Params) error {
	missedBlocks := make([]types.ValidatorMissedBlocks, 0)
	iterateValidatorSigningInfos(ctx, cdc, store, func(addr sdk.ConsAddress, info types.ValidatorSigningInfo) (stop bool) {
		bechAddr := addr.String()
		localMissedBlocks := GetValidatorMissedBlocks(ctx, cdc, store, addr, params)

		missedBlocks = append(missedBlocks, types.ValidatorMissedBlocks{
			Address:      bechAddr,
			MissedBlocks: localMissedBlocks,
		})

		return false
	})

	for _, mb := range missedBlocks {
		addr, err := sdk.ConsAddressFromBech32(mb.Address)
		if err != nil {
			return err
		}

		deleteValidatorMissedBlockBitArray(ctx, store, addr)

		for _, b := range mb.MissedBlocks {
			if b.Missed {
				if err := setMissedBlockBitmapValue(ctx, store, addr, b.Index, true); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func iterateValidatorSigningInfos(
	ctx sdk.Context,
	cdc codec.BinaryCodec,
	store storetypes.KVStore,
	cb func(address sdk.ConsAddress, info types.ValidatorSigningInfo) (stop bool),
) {
	iter := storetypes.KVStorePrefixIterator(store, ValidatorSigningInfoKeyPrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		address := ValidatorSigningInfoAddress(iter.Key())
		var info types.ValidatorSigningInfo
		cdc.MustUnmarshal(iter.Value(), &info)

		if cb(address, info) {
			break
		}
	}
}

func iterateValidatorMissedBlockBitArray(
	ctx sdk.Context,
	cdc codec.BinaryCodec,
	store storetypes.KVStore,
	addr sdk.ConsAddress,
	params types.Params,
	cb func(index int64, missed bool) (stop bool),
) {
	var index int64

	for ; index < params.SignedBlocksWindow; index++ {
		var missed gogotypes.BoolValue
		bz := store.Get(ValidatorMissedBlockBitArrayKey(addr, index))
		if bz == nil {
			continue
		}

		cdc.MustUnmarshal(bz, &missed)
		if cb(index, missed.Value) {
			break
		}
	}
}

func GetValidatorMissedBlocks(
	ctx sdk.Context,
	cdc codec.BinaryCodec,
	store storetypes.KVStore,
	addr sdk.ConsAddress,
	params types.Params,
) []types.MissedBlock {
	missedBlocks := []types.MissedBlock{}
	iterateValidatorMissedBlockBitArray(ctx, cdc, store, addr, params, func(index int64, missed bool) (stop bool) {
		missedBlocks = append(missedBlocks, types.NewMissedBlock(index, missed))
		return false
	})

	return missedBlocks
}

func deleteValidatorMissedBlockBitArray(ctx sdk.Context, store storetypes.KVStore, addr sdk.ConsAddress) {
	iter := storetypes.KVStorePrefixIterator(store, validatorMissedBlockBitArrayPrefixKey(addr))
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		store.Delete(iter.Key())
	}
}

func setMissedBlockBitmapValue(ctx sdk.Context, store storetypes.KVStore, addr sdk.ConsAddress, index int64, missed bool) error {
	// get the chunk or "word" in the logical bitmap
	chunkIndex := index / missedBlockBitmapChunkSize
	key := ValidatorMissedBlockBitmapKey(addr, chunkIndex)

	bs := bitset.New(uint(missedBlockBitmapChunkSize))
	chunk := store.Get(key)
	if chunk != nil {
		if err := bs.UnmarshalBinary(chunk); err != nil {
			return errors.Wrapf(err, "failed to decode bitmap chunk; index: %d", index)
		}
	}

	// get the bit position in the chunk of the logical bitmap
	bitIndex := uint(index % missedBlockBitmapChunkSize)
	if missed {
		bs.Set(bitIndex)
	} else {
		bs.Clear(bitIndex)
	}

	updatedChunk, err := bs.MarshalBinary()
	if err != nil {
		return errors.Wrapf(err, "failed to encode bitmap chunk; index: %d", index)
	}

	store.Set(key, updatedChunk)
	return nil
}
