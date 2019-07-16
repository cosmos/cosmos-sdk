package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/nft/internal/types"
)

// SwapOwners swaps the owners of a NFT ID
func (k Keeper) SwapOwners(ctx sdk.Context, denom string, id string, oldAddress sdk.AccAddress, newAddress sdk.AccAddress) (err sdk.Error) {
	oldOwnerIDCollection, found := k.GetOwnerByDenom(ctx, oldAddress, denom)
	if !found {
		return types.ErrUnknownCollection(types.DefaultCodespace,
			fmt.Sprintf("id collection %s doesn't exist for owner %s", denom, oldAddress),
		)
	}
	oldOwnerIDCollection, err = oldOwnerIDCollection.DeleteID(id)
	if err != nil {
		return err
	}
	k.SetOwnerByDenom(ctx, oldAddress, denom, oldOwnerIDCollection.IDs)

	newOwnerIDCollection, _ := k.GetOwnerByDenom(ctx, newAddress, denom)
	newOwnerIDCollection.AddID(id)
	k.SetOwnerByDenom(ctx, newAddress, denom, newOwnerIDCollection.IDs)
	return nil
}

// IterateCollections iterates over collections and performs a function
func (k Keeper) IterateCollections(ctx sdk.Context, handler func(collection types.Collection) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, CollectionsKeyPrefix)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var collection types.Collection
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iterator.Value(), &collection)
		if handler(collection) {
			break
		}
	}
}

// SetCollection sets the entire collection of a single denom
func (k Keeper) SetCollection(ctx sdk.Context, denom string, collection types.Collection) {
	store := ctx.KVStore(k.storeKey)
	collectionKey := GetCollectionKey(denom)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(collection)
	store.Set(collectionKey, bz)
}

// GetCollection returns a collection of NFTs
func (k Keeper) GetCollection(ctx sdk.Context, denom string) (collection types.Collection, found bool) {
	store := ctx.KVStore(k.storeKey)
	collectionKey := GetCollectionKey(denom)
	bz := store.Get(collectionKey)
	if bz == nil {
		return
	}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &collection)
	return collection, true
}

// GetCollections returns all the NFTs collections
func (k Keeper) GetCollections(ctx sdk.Context) (collections []types.Collection) {
	k.IterateCollections(ctx,
		func(collection types.Collection) (stop bool) {
			collections = append(collections, collection)
			return false
		},
	)
	return
}

// GetDenoms returns all the NFT denoms
func (k Keeper) GetDenoms(ctx sdk.Context) (denoms []string) {
	k.IterateCollections(ctx,
		func(collection types.Collection) (stop bool) {
			denoms = append(denoms, collection.Denom)
			return false
		},
	)
	return
}