package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/nft"
)

// NewClass defines a method for create a new nft class
func (k Keeper) NewClass(ctx sdk.Context, class nft.Class) error {
	if k.hasClass(ctx, class.ID) {
		return sdkerrors.Wrap(nft.ErrClassExists, class.ID)
	}
	bz := k.cdc.MustMarshal(&class)
	store := ctx.KVStore(k.storeKey)
	store.Set(classStoreKey(class.ID), bz)
	return nil
}

// UpdateClass defines a method for update a exist nft class
func (k Keeper) UpdateClass(ctx sdk.Context, class nft.Class) error {
	if !k.hasClass(ctx, class.ID) {
		return sdkerrors.Wrap(nft.ErrClassNotExists, class.ID)
	}
	bz := k.cdc.MustMarshal(&class)
	store := ctx.KVStore(k.storeKey)
	store.Set(classStoreKey(class.ID), bz)
	return nil
}

// GetClass defines a method for returning the class information of the specified id
func (k Keeper) GetClass(ctx sdk.Context, classID string) (nft.Class, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(classStoreKey(classID))

	var class nft.Class
	if len(bz) == 0 {
		return class, false
	}
	k.cdc.MustUnmarshal(bz, &class)
	return class, true
}

// GetClasses defines a method for returning all classes information
func (k Keeper) GetClasses(ctx sdk.Context) (classes []nft.Class) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, ClassKey)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var class nft.Class
		k.cdc.MustUnmarshal(iterator.Value(), &class)
		classes = append(classes, class)
	}
	return
}

func (k Keeper) hasClass(ctx sdk.Context, classID string) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(classStoreKey(classID))
}
