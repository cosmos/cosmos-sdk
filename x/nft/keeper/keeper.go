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
func NewKeeper(cdc *codec.Codec, storeKey sdk.StoreKey) Keeper {
	return Keeper{
		storeKey: storeKey,
		cdc:      cdc,
	}
}

// IsNFT returns whether an NFT exists
func (k Keeper) IsNFT(ctx sdk.Context, denom string, id uint64) (exists bool) {
	_, err := k.GetNFT(ctx, denom, id)
	return err == nil
}

// GetNFT gets the entire NFT metadata struct for a uint64
func (k Keeper) GetNFT(ctx sdk.Context, denom string, id uint64,
) (nft types.NFT, err sdk.Error) {
	collection, found := k.GetCollection(ctx, denom)
	if !found {
		return nil, types.ErrUnknownCollection(types.DefaultCodespace, fmt.Sprintf("collection of %s doesn't exist", denom))
	}
	nft, err = collection.GetNFT(id)

	if err != nil {
		return nil, err
	}
	return nft, err
}

// SetNFT sets an NFT into the store
func (k Keeper) SetNFT(ctx sdk.Context, denom string, nft types.NFT) (err sdk.Error) {
	var collection types.Collection
	collection, found := k.GetCollection(ctx, denom)
	if found {
		collection.AddNFT(nft)
	} else {
		collection = types.NewCollection(denom, types.NewNFTs(nft))
	}
	k.SetCollection(ctx, denom, collection)

	return
}

// DeleteNFT deletes an existing NFT from store
func (k Keeper) DeleteNFT(ctx sdk.Context, denom string, id uint64) (err sdk.Error) {
	collection, found := k.GetCollection(ctx, denom)
	if !found {
		return types.ErrUnknownCollection(types.DefaultCodespace, fmt.Sprintf("collection of %s doesn't exist", denom))
	}
	nft, err := collection.GetNFT(id)
	if err != nil {
		return err
	}

	err = collection.DeleteNFT(nft)
	if err != nil {
		return err
	}
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
func (k Keeper) GetCollection(ctx sdk.Context, denom string) (collection types.Collection, found bool) {
	fmt.Println("GetCollection", denom)
	store := ctx.KVStore(k.storeKey)
	collectionKey := GetCollectionKey(denom)
	bz := store.Get(collectionKey)
	if bz == nil {
		return
	}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &collection)
	return collection, true
}

// SetCollection sets the entire collection of a single denom
func (k Keeper) SetCollection(ctx sdk.Context, denom string, collection types.Collection) {
	fmt.Println("SetCollection", denom, collection)
	store := ctx.KVStore(k.storeKey)
	collectionKey := GetCollectionKey(denom)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(collection)
	store.Set(collectionKey, bz)
}

// IterateBalances iterates over the owners' balances of NFTs and performs a function
func (k Keeper) IterateBalances(ctx sdk.Context, prefix []byte, handler func(owner sdk.AccAddress, collection types.Collection) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var collection types.Collection

		owner, _ := SplitBalanceKey(iterator.Key())
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iterator.Value(), &collection)

		if handler(owner, collection) {
			break
		}
	}
}

// GetBalances returns all the NFT balances
func (k Keeper) GetBalances(ctx sdk.Context) (balances []types.Balance) {
	k.IterateBalances(ctx, BalancesKeyPrefix,
		func(owner sdk.AccAddress, collection types.Collection) (stop bool) {
			balances = append(balances, types.NewBalance(collection, owner))
			return false
		},
	)
	return
}

// GetOwnerBalances gets the all the collections of NFTs owned by an address
func (k Keeper) GetOwnerBalances(ctx sdk.Context, owner sdk.AccAddress) (collections types.Collections) {
	k.IterateBalances(ctx, GetBalancesKey(owner),
		func(_ sdk.AccAddress, collection types.Collection) (stop bool) {
			collections = append(collections, collection)
			return false
		},
	)
	return
}

// GetBalance gets the collection of NFTs owned by an address
func (k Keeper) GetBalance(ctx sdk.Context, owner sdk.AccAddress, denom string) (collection types.Collection, found bool) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(GetBalanceKey(owner, denom))
	if b == nil {
		return types.Collection{}, false
	}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(b, collection)
	return collection, true
}

// SetBalance sets a collection of NFTs owned by an address
func (k Keeper) SetBalance(ctx sdk.Context, owner sdk.AccAddress, collection types.Collection) {
	store := ctx.KVStore(k.storeKey)
	key := GetBalanceKey(owner, collection.Denom)
	store.Set(key, k.cdc.MustMarshalBinaryLengthPrefixed(collection))
}
