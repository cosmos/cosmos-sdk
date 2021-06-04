package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/nft/types"
)

// Keeper of the nft store
type Keeper struct {
	cdc      codec.BinaryMarshaler
	storeKey sdk.StoreKey
}

// NewKeeper creates a new nft Keeper instance
func NewKeeper(cdc codec.BinaryMarshaler, key sdk.StoreKey) Keeper {
	return Keeper{
		cdc:      cdc,
		storeKey: key,
	}
}

// SetNFT set the nft to the store
func (k Keeper) SetNFT(ctx sdk.Context, nft types.NFT) {
	owner, _ := sdk.AccAddressFromBech32(nft.Owner)
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryBare(&nft)
	store.Set(types.GetNFTKey(nft.Id), bz)

	ownerStore := k.getOwnerStore(ctx, owner)
	ownerStore.Set(types.MarshalNFTID(nft.Id), types.MarshalNFTID(nft.Id))
}

// GetNFT returns the nft with a given id.
func (k Keeper) GetNFT(ctx sdk.Context, id string) (nft types.NFT, has bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetNFTKey(id))
	if len(bz) == 0 {
		return nft, false
	}
	k.cdc.MustUnmarshalBinaryBare(bz, &nft)
	return nft, false
}

// IteratorNFTsByOwner iterator all the nft with a given owner.
func (k Keeper) IteratorNFTsByOwner(ctx sdk.Context, owner sdk.AccAddress, fn func(nft types.NFT)) {
	ownerStore := k.getOwnerStore(ctx, owner)
	it := ownerStore.Iterator(nil, nil)
	defer it.Close()
	for ; it.Valid(); it.Next() {
		id := string(it.Value())
		if nft, has := k.GetNFT(ctx, id); has {
			fn(nft)
		}
	}
}

// TransferOwnership transfer the ownership of an NFT to another address
func (k Keeper) TransferOwnership(ctx sdk.Context, id string,
	currentOwner, newOwner sdk.AccAddress) error {
	nft, has := k.GetNFT(ctx, id)
	if !has {
		return sdkerrors.Wrapf(types.ErrNoNFTFound, "%s", id)
	}

	// check the ownership
	if currentOwner.String() != nft.Owner {
		return sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "the owner of the nft is %s", nft.Owner)
	}

	idBytes := types.MarshalNFTID(nft.Id)
	// remove nft from current owner store
	currentOwnerStore := k.getOwnerStore(ctx, currentOwner)
	currentOwnerStore.Delete(idBytes)

	nft.Owner = newOwner.String()
	newOwnerStore := k.getOwnerStore(ctx, newOwner)
	newOwnerStore.Set(idBytes, idBytes)
	return nil
}

// getOwnerStore gets the account store of the given address.
func (k Keeper) getOwnerStore(ctx sdk.Context, owner sdk.AccAddress) prefix.Store {
	store := ctx.KVStore(k.storeKey)
	return prefix.NewStore(store, types.GetNFTByOwnerKey(owner))
}
