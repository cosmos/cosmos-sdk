package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/nft/exported"
	"github.com/cosmos/cosmos-sdk/x/nft/internal/types"
)

// IsNFT returns whether an NFT exists
func (k Keeper) IsNFT(ctx sdk.Context, denom, id string) (exists bool) {
	_, err := k.GetNFT(ctx, denom, id)
	return err == nil
}

// GetNFT gets the entire NFT metadata struct for a uint64
func (k Keeper) GetNFT(ctx sdk.Context, denom, id string) (nft exported.NFT, err sdk.Error) {
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
func (k Keeper) UpdateNFT(ctx sdk.Context, denom string, nft exported.NFT) (err sdk.Error) {
	collection, found := k.GetCollection(ctx, denom)
	if !found {
		return types.ErrUnknownCollection(types.DefaultCodespace,
			fmt.Sprintf("collection #%s doesn't exist", denom),
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
func (k Keeper) MintNFT(ctx sdk.Context, denom string, nft exported.NFT) (err sdk.Error) {
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
			fmt.Sprintf("id collection #%s doesn't exist for owner %s", denom, nft.GetOwner()),
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
