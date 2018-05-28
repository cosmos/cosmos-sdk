package slashing

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) getValidatorSigningInfo(ctx sdk.Context, address sdk.Address) (info validatorSigningInfo, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(validatorSigningInfoKey(address))
	if bz == nil {
		found = false
	} else {
		k.cdc.MustUnmarshalBinary(bz, &info)
		found = true
	}
	return
}

func (k Keeper) setValidatorSigningInfo(ctx sdk.Context, address sdk.Address, info validatorSigningInfo) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinary(info)
	store.Set(validatorSigningInfoKey(address), bz)
}

func (k Keeper) getValidatorSigningBitArray(ctx sdk.Context, address sdk.Address, index int64) (signed bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(validatorSigningBitArrayKey(address, index))
	if bz == nil {
		// lazy: treat empty key as unsigned
		signed = false
	} else {
		k.cdc.MustUnmarshalBinary(bz, &signed)
	}
	return
}

func (k Keeper) setValidatorSigningBitArray(ctx sdk.Context, address sdk.Address, index int64, signed bool) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinary(signed)
	store.Set(validatorSigningBitArrayKey(address, index), bz)
}

type validatorSigningInfo struct {
	StartHeight         int64 `json:"start_height"`
	IndexOffset         int64 `json:"index_offset"`
	JailedUntil         int64 `json:"jailed_until"`
	SignedBlocksCounter int64 `json:"signed_blocks_counter"`
}

func validatorSigningInfoKey(v sdk.Address) []byte {
	return append([]byte{0x01}, v.Bytes()...)
}

func validatorSigningBitArrayKey(v sdk.Address, i int64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(i))
	return append([]byte{0x02}, append(v.Bytes(), b...)...)
}
