package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/bank"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/nft/types"
)

// Keeper maintains the link to data storage and exposes getter/setter methods for the various parts of the state machine
type Keeper struct {
	bk bank.Keeper

	storeKey sdk.StoreKey // Unexposed key to access store from sdk.Context

	cdc *codec.Codec // The wire codec for binary encoding/decoding.
}

// NewKeeper creates new instances of the nft Keeper
func NewKeeper(coinKeeper bank.Keeper, storeKey sdk.StoreKey, cdc *codec.Codec) Keeper {
	return Keeper{
		bk:       coinKeeper,
		storeKey: storeKey,
		cdc:      cdc,
	}
}

// IsNFT returns whether an NFT exists
func (k Keeper) IsNFT(ctx sdk.Context, denom types.Denom, id types.TokenID) (exists bool) {
	_, error := k.GetNFT(ctx, denom, id)
	return error == nil
}

// GetNFT gets the entire NFT metadata struct for a types.TokenID
func (k Keeper) GetNFT(ctx sdk.Context, denom types.Denom, id types.TokenID,
) (nft types.NFT, err sdk.Error) {

	collection, found := k.GetCollection(ctx, denom)
	if !found {
		return types.NFT{}, types.ErrUnknownCollection(types.DefaultCodespace, fmt.Sprintf("collection of %s doesn't exist", denom))
	}
	nft, err = collection.GetNFT(denom, id)
	if err != nil {
		return
	}
	return
}

// SetNFT sets an NFT into the store
func (k Keeper) SetNFT(ctx sdk.Context, denom types.Denom, id types.TokenID, nft types.NFT) (err sdk.Error) {
	collection, found := k.GetCollection(ctx, denom)
	if !found {
		return types.ErrUnknownCollection(types.DefaultCodespace, fmt.Sprintf("collection of %s doesn't exist", denom))
	}
	collection[id] = nft
	k.SetCollection(ctx, denom, collection)
	return
}

// BurnNFT deletes an existing NFT from store
func (k Keeper) BurnNFT(ctx sdk.Context, denom types.Denom, id types.TokenID) (err sdk.Error) {
	collection, found := k.GetCollection(ctx, denom)
	if !found {
		return types.ErrUnknownCollection(types.DefaultCodespace, fmt.Sprintf("collection of %s doesn't exist", denom))
	}
	delete(collection, id)
	k.SetCollection(ctx, denom, collection)
	return
}

// GetCollections returns all the NFTs collections
func (k Keeper) GetCollections(ctx sdk.Context) (collections map[types.Denom]types.Collection) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, collectionKeyPrefix)
	defer iterator.Close()

	var collection types.Collection
	var denom types.Denom
	for ; iterator.Valid(); iterator.Next() {
		err := k.cdc.UnmarshalBinaryLengthPrefixed(iterator.Value(), &collection)
		if err != nil {
			panic(err)
		}

		err = k.cdc.UnmarshalBinaryLengthPrefixed(iterator.Key(), &denom)
		if err != nil {
			panic(err)
		}

		collections[denom] = collection
	}
	return
}

// GetCollection returns a collection of NFTs
func (k Keeper) GetCollection(ctx sdk.Context, denom types.Denom,
) (collection types.Collection, found bool) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get([]byte(denom))
	if b == nil {
		return nil, false
	}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(b, collection)
	return collection, true
}

// SetCollection sets the entire collection of a single denom
func (k Keeper) SetCollection(ctx sdk.Context, denom types.Denom, collection types.Collection) {
	store := ctx.KVStore(k.storeKey)
	collectionKey := GetCollectionKey(denom)
	store.Set(collectionKey, k.cdc.MustMarshalBinaryBare(collection))
}

// AddToOwner adds an NFT to owner
func (k Keeper) AddToOwner(ctx sdk.Context, denom types.Denom, id types.TokenID, nft types.NFT) {
	owner, found := k.GetOwner(ctx, nft.Owner)
	if !found {
		owner = types.NewOwner()
	}
	owner[denom] = append(owner[denom], id)
	k.SetOwner(ctx, nft.Owner, owner)
}

// SetOwner sets an owner
func (k Keeper) SetOwner(ctx sdk.Context, address sdk.AccAddress, owner types.Owner) {
	store := ctx.KVStore(k.storeKey)
	ownerKey := GetOwnerKey(address)
	store.Set(ownerKey, k.cdc.MustMarshalBinaryBare(owner))
}

// GetOwner returns a owner
func (k Keeper) GetOwner(ctx sdk.Context, address sdk.AccAddress,
) (owner types.Owner, found bool) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(GetOwnerKey(address))
	if b == nil {
		return nil, false
	}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(b, owner)
	return owner, true
}

// GetOwners returns all the NFTs owners
func (k Keeper) GetOwners(ctx sdk.Context) (owners map[string]types.Owner) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, ownerKeyPrefix)
	defer iterator.Close()

	var owner types.Owner
	var address string
	for ; iterator.Valid(); iterator.Next() {
		err := k.cdc.UnmarshalBinaryLengthPrefixed(iterator.Value(), &owner)
		if err != nil {
			panic(err)
		}

		err = k.cdc.UnmarshalBinaryLengthPrefixed(iterator.Key(), &address)
		if err != nil {
			panic(err)
		}
		owners[address] = owner
	}
	return
}
