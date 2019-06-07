package slashing

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
)

// Stored by *validator* address (not operator address)
func (k Keeper) getValidatorSigningInfo(ctx sdk.Context, address sdk.ConsAddress) (info types.ValidatorSigningInfo, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetValidatorSigningInfoKey(address))
	if bz == nil {
		found = false
		return
	}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &info)
	found = true
	return
}

// Stored by *validator* address (not operator address)
func (k Keeper) IterateValidatorSigningInfos(ctx sdk.Context,
	handler func(address sdk.ConsAddress, info types.ValidatorSigningInfo) (stop bool)) {

	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.ValidatorSigningInfoKey)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		address := types.GetValidatorSigningInfoAddress(iter.Key())
		var info types.ValidatorSigningInfo
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iter.Value(), &info)
		if handler(address, info) {
			break
		}
	}
}

// Stored by *validator* address (not operator address)
func (k Keeper) SetValidatorSigningInfo(ctx sdk.Context, address sdk.ConsAddress, info types.ValidatorSigningInfo) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(info)
	store.Set(types.GetValidatorSigningInfoKey(address), bz)
}

// Stored by *validator* address (not operator address)
func (k Keeper) getValidatorMissedBlockBitArray(ctx sdk.Context, address sdk.ConsAddress, index int64) (missed bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetValidatorMissedBlockBitArrayKey(address, index))
	if bz == nil {
		// lazy: treat empty key as not missed
		missed = false
		return
	}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &missed)
	return
}

// Stored by *validator* address (not operator address)
func (k Keeper) IterateValidatorMissedBlockBitArray(ctx sdk.Context,
	address sdk.ConsAddress, handler func(index int64, missed bool) (stop bool)) {

	store := ctx.KVStore(k.storeKey)
	index := int64(0)
	// Array may be sparse
	for ; index < k.SignedBlocksWindow(ctx); index++ {
		var missed bool
		bz := store.Get(types.GetValidatorMissedBlockBitArrayKey(address, index))
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
	store.Set(types.GetValidatorMissedBlockBitArrayKey(address, index), bz)
}

// Stored by *validator* address (not operator address)
func (k Keeper) clearValidatorMissedBlockBitArray(ctx sdk.Context, address sdk.ConsAddress) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.GetValidatorMissedBlockBitArrayPrefixKey(address))
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		store.Delete(iter.Key())
	}
}
