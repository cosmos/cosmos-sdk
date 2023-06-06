package keeper

import (
	"context"

	"cosmossdk.io/errors"
	"cosmossdk.io/store/prefix"
	"cosmossdk.io/x/nft"

	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Mint defines a method for minting a new nft
func (k Keeper) Mint(ctx context.Context, token nft.NFT, receiver sdk.AccAddress) error {
	if !k.HasClass(ctx, token.ClassId) {
		return errors.Wrap(nft.ErrClassNotExists, token.ClassId)
	}

	if k.HasNFT(ctx, token.ClassId, token.Id) {
		return errors.Wrap(nft.ErrNFTExists, token.Id)
	}

	k.mintWithNoCheck(ctx, token, receiver)
	return nil
}

// mintWithNoCheck defines a method for minting a new nft
// Note: this method does not check whether the class already exists in nft.
// The upper-layer application needs to check it when it needs to use it.
func (k Keeper) mintWithNoCheck(ctx context.Context, token nft.NFT, receiver sdk.AccAddress) {
	k.setNFT(ctx, token)
	k.setOwner(ctx, token.ClassId, token.Id, receiver)
	k.incrTotalSupply(ctx, token.ClassId)

	err := sdk.UnwrapSDKContext(ctx).EventManager().EmitTypedEvent(&nft.EventMint{
		ClassId: token.ClassId,
		Id:      token.Id,
		Owner:   receiver.String(),
	})
	if err != nil {
		panic(err)
	}
}

// Burn defines a method for burning a nft from a specific account.
// Note: When the upper module uses this method, it needs to authenticate nft
func (k Keeper) Burn(ctx context.Context, classID, nftID string) error {
	if !k.HasClass(ctx, classID) {
		return errors.Wrap(nft.ErrClassNotExists, classID)
	}

	if !k.HasNFT(ctx, classID, nftID) {
		return errors.Wrap(nft.ErrNFTNotExists, nftID)
	}

	err := k.burnWithNoCheck(ctx, classID, nftID)
	if err != nil {
		return err
	}
	return nil
}

// burnWithNoCheck defines a method for burning a nft from a specific account.
// Note: this method does not check whether the class already exists in nft.
// The upper-layer application needs to check it when it needs to use it
func (k Keeper) burnWithNoCheck(ctx context.Context, classID, nftID string) error {
	owner := k.GetOwner(ctx, classID, nftID)
	nftStore := k.getNFTStore(ctx, classID)
	nftStore.Delete([]byte(nftID))

	k.deleteOwner(ctx, classID, nftID, owner)
	k.decrTotalSupply(ctx, classID)
	err := sdk.UnwrapSDKContext(ctx).EventManager().EmitTypedEvent(&nft.EventBurn{
		ClassId: classID,
		Id:      nftID,
		Owner:   owner.String(),
	})
	if err != nil {
		return err
	}
	return nil
}

// Update defines a method for updating an exist nft
// Note: When the upper module uses this method, it needs to authenticate nft
func (k Keeper) Update(ctx context.Context, token nft.NFT) error {
	if !k.HasClass(ctx, token.ClassId) {
		return errors.Wrap(nft.ErrClassNotExists, token.ClassId)
	}

	if !k.HasNFT(ctx, token.ClassId, token.Id) {
		return errors.Wrap(nft.ErrNFTNotExists, token.Id)
	}
	k.updateWithNoCheck(ctx, token)
	return nil
}

// Update defines a method for updating an exist nft
// Note: this method does not check whether the class already exists in nft.
// The upper-layer application needs to check it when it needs to use it
func (k Keeper) updateWithNoCheck(ctx context.Context, token nft.NFT) {
	k.setNFT(ctx, token)
}

// Transfer defines a method for sending a nft from one account to another account.
// Note: When the upper module uses this method, it needs to authenticate nft
func (k Keeper) Transfer(ctx context.Context,
	classID string,
	nftID string,
	receiver sdk.AccAddress,
) error {
	if !k.HasClass(ctx, classID) {
		return errors.Wrap(nft.ErrClassNotExists, classID)
	}

	if !k.HasNFT(ctx, classID, nftID) {
		return errors.Wrap(nft.ErrNFTNotExists, nftID)
	}

	err := k.transferWithNoCheck(ctx, classID, nftID, receiver)
	if err != nil {
		return err
	}
	return nil
}

// Transfer defines a method for sending a nft from one account to another account.
// Note: this method does not check whether the class already exists in nft.
// The upper-layer application needs to check it when it needs to use it
func (k Keeper) transferWithNoCheck(ctx context.Context,
	classID string,
	nftID string,
	receiver sdk.AccAddress,
) error {
	owner := k.GetOwner(ctx, classID, nftID)
	k.deleteOwner(ctx, classID, nftID, owner)
	k.setOwner(ctx, classID, nftID, receiver)
	return nil
}

// GetNFT returns the nft information of the specified classID and nftID
func (k Keeper) GetNFT(ctx context.Context, classID, nftID string) (nft.NFT, bool) {
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
func (k Keeper) GetNFTsOfClassByOwner(ctx context.Context, classID string, owner sdk.AccAddress) (nfts []nft.NFT) {
	ownerStore := k.getClassStoreByOwner(ctx, owner, classID)
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
func (k Keeper) GetNFTsOfClass(ctx context.Context, classID string) (nfts []nft.NFT) {
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
func (k Keeper) GetOwner(ctx context.Context, classID, nftID string) sdk.AccAddress {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(ownerStoreKey(classID, nftID))
	if err != nil {
		panic(err)
	}
	return sdk.AccAddress(bz)
}

// GetBalance returns the specified account, the number of all nfts under the specified classID
func (k Keeper) GetBalance(ctx context.Context, classID string, owner sdk.AccAddress) uint64 {
	nfts := k.GetNFTsOfClassByOwner(ctx, classID, owner)
	return uint64(len(nfts))
}

// GetTotalSupply returns the number of all nfts under the specified classID
func (k Keeper) GetTotalSupply(ctx context.Context, classID string) uint64 {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(classTotalSupply(classID))
	if err != nil {
		panic(err)
	}
	return sdk.BigEndianToUint64(bz)
}

// HasNFT determines whether the specified classID and nftID exist
func (k Keeper) HasNFT(ctx context.Context, classID, id string) bool {
	store := k.getNFTStore(ctx, classID)
	return store.Has([]byte(id))
}

func (k Keeper) setNFT(ctx context.Context, token nft.NFT) {
	nftStore := k.getNFTStore(ctx, token.ClassId)
	bz := k.cdc.MustMarshal(&token)
	nftStore.Set([]byte(token.Id), bz)
}

func (k Keeper) setOwner(ctx context.Context, classID, nftID string, owner sdk.AccAddress) {
	store := k.storeService.OpenKVStore(ctx)
	err := store.Set(ownerStoreKey(classID, nftID), owner.Bytes())
	if err != nil {
		panic(err)
	}

	ownerStore := k.getClassStoreByOwner(ctx, owner, classID)
	ownerStore.Set([]byte(nftID), Placeholder)
}

func (k Keeper) deleteOwner(ctx context.Context, classID, nftID string, owner sdk.AccAddress) {
	store := k.storeService.OpenKVStore(ctx)
	err := store.Delete(ownerStoreKey(classID, nftID))
	if err != nil {
		panic(err)
	}
	ownerStore := k.getClassStoreByOwner(ctx, owner, classID)
	ownerStore.Delete([]byte(nftID))
}

func (k Keeper) getNFTStore(ctx context.Context, classID string) prefix.Store {
	store := k.storeService.OpenKVStore(ctx)
	return prefix.NewStore(runtime.KVStoreAdapter(store), nftStoreKey(classID))
}

func (k Keeper) getClassStoreByOwner(ctx context.Context, owner sdk.AccAddress, classID string) prefix.Store {
	store := k.storeService.OpenKVStore(ctx)
	key := nftOfClassByOwnerStoreKey(owner, classID)
	return prefix.NewStore(runtime.KVStoreAdapter(store), key)
}

func (k Keeper) prefixStoreNftOfClassByOwner(ctx context.Context, owner sdk.AccAddress) prefix.Store {
	store := k.storeService.OpenKVStore(ctx)
	key := prefixNftOfClassByOwnerStoreKey(owner)
	return prefix.NewStore(runtime.KVStoreAdapter(store), key)
}

func (k Keeper) incrTotalSupply(ctx context.Context, classID string) {
	supply := k.GetTotalSupply(ctx, classID) + 1
	k.updateTotalSupply(ctx, classID, supply)
}

func (k Keeper) decrTotalSupply(ctx context.Context, classID string) {
	supply := k.GetTotalSupply(ctx, classID) - 1
	k.updateTotalSupply(ctx, classID, supply)
}

func (k Keeper) updateTotalSupply(ctx context.Context, classID string, supply uint64) {
	store := k.storeService.OpenKVStore(ctx)
	supplyKey := classTotalSupply(classID)
	err := store.Set(supplyKey, sdk.Uint64ToBigEndian(supply))
	if err != nil {
		panic(err)
	}
}
