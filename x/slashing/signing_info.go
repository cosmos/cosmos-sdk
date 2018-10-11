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
	k.cdc.MustUnmarshalBinary(bz, &info)
	found = true
	return
}

// Stored by *validator* address (not operator address)
func (k Keeper) setValidatorSigningInfo(ctx sdk.Context, address sdk.ConsAddress, info ValidatorSigningInfo) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinary(info)
	store.Set(GetValidatorSigningInfoKey(address), bz)
}

// Stored by *validator* address (not operator address)
func (k Keeper) getValidatorSigningBitArray(ctx sdk.Context, address sdk.ConsAddress, index int64) (signed bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(GetValidatorSigningBitArrayKey(address, index))
	if bz == nil {
		// lazy: treat empty key as unsigned
		signed = false
		return
	}
	k.cdc.MustUnmarshalBinary(bz, &signed)
	return
}

// Stored by *validator* address (not operator address)
func (k Keeper) setValidatorSigningBitArray(ctx sdk.Context, address sdk.ConsAddress, index int64, signed bool) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinary(signed)
	store.Set(GetValidatorSigningBitArrayKey(address, index), bz)
}

// Construct a new `ValidatorSigningInfo` struct
func NewValidatorSigningInfo(startHeight int64, indexOffset int64, jailedUntil time.Time, missedBlocksCounter int64) ValidatorSigningInfo {
	return ValidatorSigningInfo{
		StartHeight:         startHeight,
		IndexOffset:         indexOffset,
		JailedUntil:         jailedUntil,
		MissedBlocksCounter: missedBlocksCounter,
	}
}

// Signing info for a validator
type ValidatorSigningInfo struct {
	StartHeight         int64     `json:"start_height"`            // height at which validator was first a candidate OR was unjailed
	IndexOffset         int64     `json:"index_offset"`            // index offset into signed block bit array
	JailedUntil         time.Time `json:"jailed_until"`            // timestamp validator cannot be unjailed until
	MissedBlocksCounter int64     `json:"misseded_blocks_counter"` // missed blocks counter (to avoid scanning the array every time)
}

// Return human readable signing info
func (i ValidatorSigningInfo) HumanReadableString() string {
	return fmt.Sprintf("Start height: %d, index offset: %d, jailed until: %v, missed blocks counter: %d",
		i.StartHeight, i.IndexOffset, i.JailedUntil, i.MissedBlocksCounter)
}
