package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/nft/types"
)

// Issue defines a method for create a new nft type
func (k Keeper) Issue(ctx sdk.Context,
	typ string,
	name string,
	symbol string,
	description string,
	mintRestricted bool,
	editRestricted bool,
	issuer sdk.AccAddress) error {
	if k.HasType(ctx, typ) {
		return sdkerrors.Wrap(types.ErrNFTTypeExists, typ)
	}
	bz := k.cdc.MustMarshal(&types.Metadata{
		Type:           typ,
		Name:           name,
		Symbol:         symbol,
		Description:    description,
		MintRestricted: mintRestricted,
		EditRestricted: editRestricted,
	})
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetTypeKey(typ), bz)
	return nil
}

func (k Keeper) GetMetadata(ctx sdk.Context, typ string) types.Metadata {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetTypeKey(typ))

	var metadata types.Metadata
	k.cdc.MustUnmarshal(bz, &metadata)
	return metadata
}

func (k Keeper) HasType(ctx sdk.Context, typ string) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.GetTypeKey(typ))
}
