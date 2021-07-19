package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/nft"
)

// Mint defines a method for minting a new nft
func (k Keeper) Mint(ctx sdk.Context, newNFT nft.NFT, minter sdk.AccAddress) error {
	if !k.HasClass(ctx, newNFT.ClassID) {
		return sdkerrors.Wrap(nft.ErrClassNotExists, newNFT.ClassID)
	}

	if k.HasNFT(ctx, newNFT.ClassID, newNFT.ID) {
		return sdkerrors.Wrap(nft.ErrNFTExists, newNFT.ID)
	}

	k.setNFT(ctx, newNFT)
	k.setOwner(ctx, newNFT.ClassID, newNFT.ID, minter)
	k.incrTotalSupply(ctx, newNFT.ClassID)
	return ctx.EventManager().EmitTypedEvent(&nft.EventMint{
		ClassID: newNFT.ClassID,
		ID:      newNFT.ID,
		Minter:  minter.String(),
	})
}

// Burn defines a method for burning a nft from a specific account.
func (k Keeper) Burn(ctx sdk.Context, classID string, nftID string) error {
	if !k.HasClass(ctx, classID) {
		return sdkerrors.Wrap(nft.ErrClassNotExists, classID)
	}

	if !k.HasNFT(ctx, classID, nftID) {
		return sdkerrors.Wrap(nft.ErrNFTNotExists, nftID)
	}

	owner := k.GetOwner(ctx, classID, nftID)
	nftStore := k.getNFTStore(ctx, classID)
	nftStore.Delete([]byte(nftID))

	k.deleteOwner(ctx, classID, nftID, owner)
	k.decrTotalSupply(ctx, classID)
	return nil
}

// Update defines a method for update a exist nft
func (k Keeper) Update(ctx sdk.Context, updNFT nft.NFT) error {
	if !k.HasClass(ctx, updNFT.ClassID) {
		return sdkerrors.Wrap(nft.ErrClassNotExists, updNFT.ClassID)
	}

	if !k.HasNFT(ctx, updNFT.ClassID, updNFT.ID) {
		return sdkerrors.Wrap(nft.ErrNFTNotExists, updNFT.ID)
	}
	k.setNFT(ctx, updNFT)
	return nil
}

// Transfer defines a method for sending a nft from one account to another account.
func (k Keeper) Transfer(ctx sdk.Context,
	classID string,
	nftID string,
	receiver sdk.AccAddress) error {
	if !k.HasClass(ctx, classID) {
		return sdkerrors.Wrap(nft.ErrClassNotExists, classID)
	}

	if !k.HasNFT(ctx, classID, nftID) {
		return sdkerrors.Wrap(nft.ErrNFTNotExists, nftID)
	}

	owner := k.GetOwner(ctx, classID, nftID)
	k.deleteOwner(ctx, classID, nftID, owner)
	k.setOwner(ctx, classID, nftID, receiver)
	return nil
}

// GetNFT returns the nft information of the specified classID and nftID
func (k Keeper) GetNFT(ctx sdk.Context, classID, nftID string) (nft.NFT, bool) {
	store := k.getNFTStore(ctx, classID)
	bz := store.Get([]byte(nftID))
	if len(bz) == 0 {
		return nft.NFT{}, false
	}
	var nft nft.NFT
	k.cdc.MustUnmarshal(bz, &nft)
	return nft, true
}

// GetNFTsOfClassByOwner returns all nft information of the specified classID under the specified owner
func (k Keeper) GetNFTsOfClassByOwner(ctx sdk.Context, classID string, owner sdk.AccAddress) (nfts []nft.NFT) {
	ownerStore := k.getOwnerStore(ctx, owner, classID)
	iterator := ownerStore.Iterator(nil, nil)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		nft, has := k.GetNFT(ctx, classID, string(iterator.Key()))
		if has {
			nfts = append(nfts, nft)
		}
	}
	return nfts
}

// GetNFTsOfClass returns all nft information under the specified classID
func (k Keeper) GetNFTsOfClass(ctx sdk.Context, classID string) (nfts []nft.NFT) {
	nftStore := k.getNFTStore(ctx, classID)
	iterator := nftStore.Iterator(nil, nil)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var nft nft.NFT
		k.cdc.MustUnmarshal(iterator.Value(), &nft)
		nfts = append(nfts, nft)
	}
	return nfts
}

// GetOwner returns the owner information of the specified nft
func (k Keeper) GetOwner(ctx sdk.Context, classID string, nftID string) sdk.AccAddress {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(ownerStoreKey(classID, nftID))
	return sdk.AccAddress(bz)
}

// GetBalance returns the specified account, the number of all nfts under the specified classID
func (k Keeper) GetBalance(ctx sdk.Context, classID string, owner sdk.AccAddress) uint64 {
	nfts := k.GetNFTsOfClassByOwner(ctx, classID, owner)
	return uint64(len(nfts))
}

// GetTotalSupply returns the number of all nfts under the specified classID
func (k Keeper) GetTotalSupply(ctx sdk.Context, classID string) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(classTotalSupply(classID))
	return sdk.BigEndianToUint64(bz)
}

// HasNFT determines whether the specified classID and nftID exist
func (k Keeper) HasNFT(ctx sdk.Context, classID, id string) bool {
	store := k.getNFTStore(ctx, classID)
	return store.Has([]byte(id))
}

func (k Keeper) setNFT(ctx sdk.Context, nft nft.NFT) {
	nftStore := k.getNFTStore(ctx, nft.ClassID)
	bz := k.cdc.MustMarshal(&nft)
	nftStore.Set([]byte(nft.ID), bz)
}

func (k Keeper) setOwner(ctx sdk.Context, classID, nftID string, owner sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Set(ownerStoreKey(classID, nftID), owner.Bytes())

	ownerStore := k.getOwnerStore(ctx, owner, classID)
	ownerStore.Set([]byte(nftID), []byte{0x01})
}

func (k Keeper) deleteOwner(ctx sdk.Context, classID, nftID string, owner sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(ownerStoreKey(classID, nftID))

	ownerStore := k.getOwnerStore(ctx, owner, classID)
	ownerStore.Delete([]byte(nftID))
}

func (k Keeper) getNFTStore(ctx sdk.Context, classID string) prefix.Store {
	store := ctx.KVStore(k.storeKey)
	return prefix.NewStore(store, nftStoreKey(classID))
}

func (k Keeper) getOwnerStore(ctx sdk.Context, owner sdk.AccAddress, classID string) prefix.Store {
	store := ctx.KVStore(k.storeKey)
	key := nftOfClassByOwnerStoreKey(owner, classID)
	return prefix.NewStore(store, key)
}

func (k Keeper) incrTotalSupply(ctx sdk.Context, classID string) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(classTotalSupply(classID))
	supply := sdk.BigEndianToUint64(bz) + 1
	store.Set(classTotalSupply(classID), sdk.Uint64ToBigEndian(supply))
}

func (k Keeper) decrTotalSupply(ctx sdk.Context, classID string) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(classTotalSupply(classID))
	supply := sdk.BigEndianToUint64(bz) - 1
	store.Set(classTotalSupply(classID), sdk.Uint64ToBigEndian(supply))
}
