package slashing

import (
	"encoding/binary"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Stored by *validator* address (not owner address)
func (k Keeper) getValidatorSigningInfo(ctx sdk.Context, address sdk.ValAddress) (info ValidatorSigningInfo, found bool) {
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

// Stored by *validator* address (not owner address)
func (k Keeper) setValidatorSigningInfo(ctx sdk.Context, address sdk.ValAddress, info ValidatorSigningInfo) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinary(info)
	store.Set(GetValidatorSigningInfoKey(address), bz)
}

// Stored by *validator* address (not owner address)
func (k Keeper) getValidatorSigningBitArray(ctx sdk.Context, address sdk.ValAddress, index int64) (signed bool) {
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

// Stored by *validator* address (not owner address)
func (k Keeper) setValidatorSigningBitArray(ctx sdk.Context, address sdk.ValAddress, index int64, signed bool) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinary(signed)
	store.Set(GetValidatorSigningBitArrayKey(address, index), bz)
}

// Construct a new `ValidatorSigningInfo` struct
func NewValidatorSigningInfo(startHeight int64, indexOffset int64, jailedUntil int64, signedBlocksCounter int64) ValidatorSigningInfo {
	return ValidatorSigningInfo{
		StartHeight:         startHeight,
		IndexOffset:         indexOffset,
		JailedUntil:         jailedUntil,
		SignedBlocksCounter: signedBlocksCounter,
	}
}

// Signing info for a validator
type ValidatorSigningInfo struct {
	StartHeight         int64 `json:"start_height"`          // height at which validator was first a candidate OR was unrevoked
	IndexOffset         int64 `json:"index_offset"`          // index offset into signed block bit array
	JailedUntil         int64 `json:"jailed_until"`          // timestamp validator cannot be unrevoked until
	SignedBlocksCounter int64 `json:"signed_blocks_counter"` // signed blocks counter (to avoid scanning the array every time)
}

// Return human readable signing info
func (i ValidatorSigningInfo) HumanReadableString() string {
	return fmt.Sprintf("Start height: %d, index offset: %d, jailed until: %d, signed blocks counter: %d",
		i.StartHeight, i.IndexOffset, i.JailedUntil, i.SignedBlocksCounter)
}

// Stored by *validator* address (not owner address)
func GetValidatorSigningInfoKey(v sdk.ValAddress) []byte {
	return append([]byte{0x01}, v.Bytes()...)
}

// Stored by *validator* address (not owner address)
func GetValidatorSigningBitArrayKey(v sdk.ValAddress, i int64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(i))
	return append([]byte{0x02}, append(v.Bytes(), b...)...)
}
