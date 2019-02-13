package slashing

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Stored by *validator* address (not operator address)
func (k Keeper) getValidatorSigningInfo(ctx sdk.Context, address sdk.ConsAddress) (info ValidatorSigningInfo, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(GetValidatorSigningInfoKey(address))
	if bz == nil {
		found = false
		return
	}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &info)
	found = true
	return
}

// Stored by *validator* address (not operator address)
func (k Keeper) IterateValidatorSigningInfos(ctx sdk.Context, handler func(address sdk.ConsAddress, info ValidatorSigningInfo) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, ValidatorSigningInfoKey)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		address := GetValidatorSigningInfoAddress(iter.Key())
		var info ValidatorSigningInfo
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iter.Value(), &info)
		if handler(address, info) {
			break
		}
	}
}

// Stored by *validator* address (not operator address)
func (k Keeper) SetValidatorSigningInfo(ctx sdk.Context, address sdk.ConsAddress, info ValidatorSigningInfo) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(info)
	store.Set(GetValidatorSigningInfoKey(address), bz)
}

// Stored by *validator* address (not operator address)
func (k Keeper) getValidatorMissedBlockBitArray(ctx sdk.Context, address sdk.ConsAddress, index int64) (missed bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(GetValidatorMissedBlockBitArrayKey(address, index))
	if bz == nil {
		// lazy: treat empty key as not missed
		missed = false
		return
	}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &missed)
	return
}

// Stored by *validator* address (not operator address)
func (k Keeper) IterateValidatorMissedBlockBitArray(ctx sdk.Context, address sdk.ConsAddress, handler func(index int64, missed bool) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	index := int64(0)
	// Array may be sparse
	for ; index < k.SignedBlocksWindow(ctx); index++ {
		var missed bool
		bz := store.Get(GetValidatorMissedBlockBitArrayKey(address, index))
		if bz == nil {
			continue
		}
		k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &missed)
		if handler(index, missed) {
			break
		}
	}
}

// Stored by *validator* address (not operator address)
func (k Keeper) setValidatorMissedBlockBitArray(ctx sdk.Context, address sdk.ConsAddress, index int64, missed bool) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(missed)
	store.Set(GetValidatorMissedBlockBitArrayKey(address, index), bz)
}

// Stored by *validator* address (not operator address)
func (k Keeper) clearValidatorMissedBlockBitArray(ctx sdk.Context, address sdk.ConsAddress) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, GetValidatorMissedBlockBitArrayPrefixKey(address))
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		store.Delete(iter.Key())
	}
}

// Signing info for a validator
type ValidatorSigningInfo struct {
	StartHeight         int64     `json:"start_height"`          // height at which validator was first a candidate OR was unjailed
	IndexOffset         int64     `json:"index_offset"`          // index offset into signed block bit array
	JailedUntil         time.Time `json:"jailed_until"`          // timestamp validator cannot be unjailed until
	Tombstoned          bool      `json:"tombstoned"`            // whether or not a validator has been tombstoned (killed out of validator set)
	MissedBlocksCounter int64     `json:"missed_blocks_counter"` // missed blocks counter (to avoid scanning the array every time)
}

// Construct a new `ValidatorSigningInfo` struct
func NewValidatorSigningInfo(startHeight, indexOffset int64, jailedUntil time.Time,
	tombstoned bool, missedBlocksCounter int64) ValidatorSigningInfo {

	return ValidatorSigningInfo{
		StartHeight:         startHeight,
		IndexOffset:         indexOffset,
		JailedUntil:         jailedUntil,
		Tombstoned:          tombstoned,
		MissedBlocksCounter: missedBlocksCounter,
	}
}

// Return human readable signing info
func (i ValidatorSigningInfo) String() string {
	return fmt.Sprintf(`Start Height:          %d
Index Offset:          %d
Jailed Until:          %v
Tombstoned:            %t
Missed Blocks Counter: %d`,
		i.StartHeight, i.IndexOffset, i.JailedUntil,
		i.Tombstoned, i.MissedBlocksCounter)
}
