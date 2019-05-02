package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/nft/types"
)

// Keeper maintains the link to data storage and exposes getter/setter methods for the various parts of the state machine
type Keeper struct {
	storeKey sdk.StoreKey // Unexposed key to access store from sdk.Context

	cdc *codec.Codec // The wire codec for binary encoding/decoding.
}

// NewKeeper creates new instances of the nft Keeper
func NewKeeper(storeKey sdk.StoreKey, cdc *codec.Codec) Keeper {
	return Keeper{
		storeKey: storeKey,
		cdc:      cdc,
	}
}

// IsNFT returns whether an NFT exists
func (k Keeper) IsNFT(ctx sdk.Context, denom string, id uint64) (exists bool) {
	_, error := k.GetNFT(ctx, denom, id)
	return error == nil
}

// GetNFT gets the entire NFT metadata struct for a uint64
func (k Keeper) GetNFT(ctx sdk.Context, denom string, id uint64,
) (nft types.NFT, err sdk.Error) {

	collection, found := k.GetCollection(ctx, denom)
	if !found {
		return types.NFT{}, types.ErrUnknownCollection(types.DefaultCodespace, fmt.Sprintf("collection of %s doesn't exist", denom))
	}
	nft, err = collection.GetNFT(id)
	if err != nil {
		return
	}
	return
}

// SetNFT sets an NFT into the store
func (k Keeper) SetNFT(ctx sdk.Context, denom string, id uint64, nft types.NFT) (err sdk.Error) {
	collection, found := k.GetCollection(ctx, denom)
	if !found {
		return types.ErrUnknownCollection(types.DefaultCodespace, fmt.Sprintf("collection of %s doesn't exist", denom))
	}

	collection.AddNFT(nft)
	k.SetCollection(ctx, denom, collection)
	return
}

// BurnNFT deletes an existing NFT from store
func (k Keeper) BurnNFT(ctx sdk.Context, denom string, id uint64) (err sdk.Error) {\
	collection, found := k.GetCollection(ctx, denom)
	if !found {
		return types.ErrUnknownCollection(types.DefaultCodespace, fmt.Sprintf("collection of %s doesn't exist", denom))
	}
	delete(collection, id)
	k.SetCollection(ctx, denom, collection)
	return
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

// GetCollection returns a collection of NFTs
func (k Keeper) GetCollection(ctx sdk.Context, denom string,
) (collection types.Collection, found bool) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get([]byte(denom))
	if b == nil {
		return collection, false
	}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(b, collection)
	return collection, true
}

// SetCollection sets the entire collection of a single denom
func (k Keeper) SetCollection(ctx sdk.Context, denom string, collection types.Collection) {
	store := ctx.KVStore(k.storeKey)
	collectionKey := GetCollectionKey(denom)
	store.Set(collectionKey, k.cdc.MustMarshalBinaryBare(collection))
}

// IterateNFTBalances iterates over the owners' balances of NFTs and performs a function
func (k Keeper) IterateNFTBalances(ctx sdk.Context, handler func(owner sdk.AccAddress, collection types.Collection) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, NFTBalancesKeyPrefix)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var collection types.Collection

		owner := GetNFTBalancesAddress(iterator.Key())
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iterator.Value(), &collection)

		if handler(owner, collection) {
			break
		}
	}
}
