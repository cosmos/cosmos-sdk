package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/nft/internal/types"
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
func (k Keeper) IsNFT(ctx sdk.Context, denom, id string) (exists bool) {
	_, err := k.GetNFT(ctx, denom, id)
	return err == nil
}

// GetNFT gets the entire NFT metadata struct for a uint64
func (k Keeper) GetNFT(ctx sdk.Context, denom, id string) (nft types.NFT, err sdk.Error) {
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

// UpdateNFT updates an already existing NFTs
func (k Keeper) UpdateNFT(ctx sdk.Context, denom string, nft types.NFT) (err sdk.Error) {
	collection, found := k.GetCollection(ctx, denom)
	if !found {
		return types.ErrUnknownCollection(types.DefaultCodespace,
			fmt.Sprintf("Collection #%s doesn't exist", denom),
		)
	}
	oldNFT, err := collection.GetNFT(nft.GetID())
	if err != nil {
		return err
	}
	// if the owner changed then update the owners KVStore too
	if !oldNFT.GetOwner().Equals(nft.GetOwner()) {
		err = k.SwapOwners(ctx, denom, nft.GetID(), oldNFT.GetOwner(), nft.GetOwner())
		if err != nil {
			return err
		}
	}
	collection, err = collection.UpdateNFT(nft)
	if err != nil {
		return err
	}
	k.SetCollection(ctx, denom, collection)
	return nil
}

// MintNFT mints an NFT and manages that NFTs existence within Collections and Owners
func (k Keeper) MintNFT(ctx sdk.Context, denom string, nft types.NFT) (err sdk.Error) {
	collection, found := k.GetCollection(ctx, denom)
	if found {
		collection, err = collection.AddNFT(nft)
		if err != nil {
			return err
		}
	} else {
		collection = types.NewCollection(denom, types.NewNFTs(nft))
	}
	k.SetCollection(ctx, denom, collection)

	ownerIDCollection, _ := k.GetOwnerByDenom(ctx, nft.GetOwner(), denom)
	ownerIDCollection = ownerIDCollection.AddID(nft.GetID())
	k.SetOwnerByDenom(ctx, nft.GetOwner(), denom, ownerIDCollection.IDs)
	return
}

// DeleteNFT deletes an existing NFT from store
func (k Keeper) DeleteNFT(ctx sdk.Context, denom, id string) (err sdk.Error) {
	collection, found := k.GetCollection(ctx, denom)
	if !found {
		return types.ErrUnknownCollection(types.DefaultCodespace, fmt.Sprintf("collection of %s doesn't exist", denom))
	}
	nft, err := collection.GetNFT(id)
	if err != nil {
		return err
	}
	ownerIDCollection, found := k.GetOwnerByDenom(ctx, nft.GetOwner(), denom)
	if !found {
		return types.ErrUnknownCollection(types.DefaultCodespace,
			fmt.Sprintf("ID Collection #%s doesn't exist for owner %s", denom, nft.GetOwner()),
		)
	}
	ownerIDCollection, err = ownerIDCollection.DeleteID(nft.GetID())
	if err != nil {
		return err
	}
	k.SetOwnerByDenom(ctx, nft.GetOwner(), denom, ownerIDCollection.IDs)

	collection, err = collection.DeleteNFT(nft)
	if err != nil {
		return err
	}

	k.SetCollection(ctx, denom, collection)

	return
}

// SwapOwners swaps the owners of a NFT ID
func (k Keeper) SwapOwners(ctx sdk.Context, denom string, id string, oldAddress sdk.AccAddress, newAddress sdk.AccAddress) (err sdk.Error) {
	oldOwnerIDCollection, found := k.GetOwnerByDenom(ctx, oldAddress, denom)
	if !found {
		return types.ErrUnknownCollection(types.DefaultCodespace,
			fmt.Sprintf("ID Collection %s doesn't exist for owner %s", denom, oldAddress),
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

// GetOwners returns all the Owners ID Collections
func (k Keeper) GetOwners(ctx sdk.Context) (owners []types.Owner) {
	var foundOwners = make(map[string]bool)
	k.IterateOwners(ctx,
		func(owner types.Owner) (stop bool) {
			if _, ok := foundOwners[owner.Address.String()]; !ok {
				foundOwners[owner.Address.String()] = true
				owners = append(owners, owner)
			}
			return false
		},
	)
	return
}

// GetOwner gets all the ID Collections owned by an address
func (k Keeper) GetOwner(ctx sdk.Context, address sdk.AccAddress) (owner types.Owner) {
	var idCollections []types.IDCollection
	k.IterateIDCollections(ctx, GetOwnersKey(address),
		func(_ sdk.AccAddress, idCollection types.IDCollection) (stop bool) {
			idCollections = append(idCollections, idCollection)
			return false
		},
	)
	return types.NewOwner(address, idCollections...)
}

// GetOwnerByDenom gets the ID Collection owned by an address of a specific denom
func (k Keeper) GetOwnerByDenom(ctx sdk.Context, owner sdk.AccAddress, denom string) (idCollection types.IDCollection, found bool) {

	store := ctx.KVStore(k.storeKey)
	b := store.Get(GetOwnerKey(owner, denom))
	if b == nil {
		return types.NewIDCollection(denom, []string{}), false
	}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(b, &idCollection)
	return idCollection, true
}

// SetOwnerByDenom sets a collection of NFT IDs owned by an address
func (k Keeper) SetOwnerByDenom(ctx sdk.Context, owner sdk.AccAddress, denom string, ids []string) {
	store := ctx.KVStore(k.storeKey)
	key := GetOwnerKey(owner, denom)

	var idCollection types.IDCollection
	idCollection.Denom = denom
	idCollection.IDs = ids

	store.Set(key, k.cdc.MustMarshalBinaryLengthPrefixed(idCollection))
}

// SetOwner sets an entire Owner
func (k Keeper) SetOwner(ctx sdk.Context, owner types.Owner) {
	for _, idCollection := range owner.IDCollections {
		k.SetOwnerByDenom(ctx, owner.Address, idCollection.Denom, idCollection.IDs)
	}
}

// SetOwners sets all Owners
func (k Keeper) SetOwners(ctx sdk.Context, owners []types.Owner) {
	for _, owner := range owners {
		k.SetOwner(ctx, owner)
	}
}

// IterateIDCollections iterates over the IDCollections by Owner and performs a function
func (k Keeper) IterateIDCollections(ctx sdk.Context, prefix []byte,
	handler func(owner sdk.AccAddress, idCollection types.IDCollection) (stop bool)) {

	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {

		var idCollection types.IDCollection
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iterator.Value(), &idCollection)

		owner, _ := SplitOwnerKey(iterator.Key())
		if handler(owner, idCollection) {
			break
		}
	}
}

// IterateOwners iterates over all Owners and performs a function
func (k Keeper) IterateOwners(ctx sdk.Context, handler func(owner types.Owner) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, OwnersKeyPrefix)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var owner types.Owner

		address, _ := SplitOwnerKey(iterator.Key())
		owner = k.GetOwner(ctx, address)

		if handler(owner) {
			break
		}
	}
}
