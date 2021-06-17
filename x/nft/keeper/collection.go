package keeper

import (
	gogotypes "github.com/gogo/protobuf/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k keeper) getTotalSupply(ctx sdk.Context, denomID string) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(KeyCollectionSupply(denomID))
	if len(bz) == 0 {
		return 0
	}
	return mustUnmarshalSupply(k.cdc, bz)
}

func (k keeper) increaseSupply(ctx sdk.Context, collectionId string) {
	supply := k.getTotalSupply(ctx, collectionId)
	supply++

	store := ctx.KVStore(k.storeKey)
	bz := mustMarshalSupply(k.cdc, supply)
	store.Set(KeyCollectionSupply(collectionId), bz)
}

func (k keeper) decreaseSupply(ctx sdk.Context, collectionId string) {
	supply := k.getTotalSupply(ctx, collectionId)
	supply--

	store := ctx.KVStore(k.storeKey)
	if supply == 0 {
		store.Delete(KeyCollectionSupply(collectionId))
		return
	}

	bz := mustMarshalSupply(k.cdc, supply)
	store.Set(KeyCollectionSupply(collectionId), bz)
}

func mustUnmarshalSupply(cdc codec.Codec, value []byte) uint64 {
	var supplyWrap gogotypes.UInt64Value
	cdc.MustUnmarshal(value, &supplyWrap)
	return supplyWrap.Value
}

func mustMarshalSupply(cdc codec.Codec, supply uint64) []byte {
	supplyWrap := gogotypes.UInt64Value{Value: supply}
	return cdc.MustMarshal(&supplyWrap)
}
