package keeper

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/nft/types"
)

// MintNFT defines a method for minting a new nft
func (k Keeper) MintNFT(ctx sdk.Context,
	typ string,
	id string,
	uri string,
	data *codectypes.Any,
	minter sdk.AccAddress) error {
	metadata, has := k.GetMetadata(ctx, typ)
	if !has {
		return sdkerrors.Wrap(types.ErrTypeNotExists, typ)
	}

	if k.HasNFT(ctx, typ, id) {
		return sdkerrors.Wrap(types.ErrNFTExists, id)
	}

	nft := types.NFT{
		Type: typ,
		ID:   id,
		URI:  uri,
		Data: data,
	}

	coin := nft.Coin()
	bkMetadata := banktypes.Metadata{
		Symbol:      metadata.Symbol,
		Base:        coin.GetDenom(),
		Name:        metadata.Name,
		Description: metadata.Description,
	}
	mintedCoins := sdk.NewCoins(coin)
	k.bk.SetDenomMetaData(ctx, bkMetadata)
	k.bk.MintCoins(ctx, types.ModuleName, mintedCoins)
	k.bk.SendCoinsFromModuleToAccount(ctx, types.ModuleName, minter, mintedCoins)

	bz := k.cdc.MustMarshal(&nft)
	nftStore := k.getNFTStore(ctx, nft.Type)
	nftStore.Set(types.GetNFTIdKey(nft.ID), bz)
	return nil
}

// EditNFT defines a method for editing a exist nft
func (k Keeper) EditNFT(ctx sdk.Context,
	typ string,
	id string,
	uri string,
	data *codectypes.Any,
	editor sdk.AccAddress) error {
	// Assert whether nft type exists
	metadata, has := k.GetMetadata(ctx, typ)
	if !has {
		return sdkerrors.Wrap(types.ErrTypeNotExists, typ)
	}

	// If nft does not allow editing, return an error
	if metadata.EditRestricted {
		return sdkerrors.Wrap(types.ErrNFTEditRestricted, id)
	}

	nft, has := k.GetNFT(ctx, typ, id)
	if !has {
		return sdkerrors.Wrap(types.ErrNFTNotExists, id)
	}
	nft.URI = uri
	nft.Data = data

	bz := k.cdc.MustMarshal(&nft)
	nftStore := k.getNFTStore(ctx, nft.Type)
	nftStore.Set(types.GetNFTIdKey(nft.ID), bz)
	return nil
}

// SendNFT defines a method for sending a nft from one account to another account.
func (k Keeper) SendNFT(ctx sdk.Context,
	typ string,
	id string,
	sender sdk.AccAddress,
	receiver sdk.AccAddress) error {
	if !k.HasType(ctx, typ) {
		return sdkerrors.Wrap(types.ErrTypeNotExists, typ)
	}

	if !k.HasNFT(ctx, typ, id) {
		return sdkerrors.Wrap(types.ErrNFTNotExists, typ)
	}
	sentCoins := sdk.NewCoins(sdk.NewCoin(types.CreateDenom(typ, id), sdk.OneInt()))
	return k.bk.SendCoins(ctx, sender, receiver, sentCoins)
}

// BurnNFT defines a method for burning a nft from a specific account.
func (k Keeper) BurnNFT(ctx sdk.Context,
	typ string,
	id string,
	destroyer sdk.AccAddress) error {
	if !k.HasType(ctx, typ) {
		return sdkerrors.Wrap(types.ErrTypeNotExists, typ)
	}

	if !k.HasNFT(ctx, typ, id) {
		return sdkerrors.Wrap(types.ErrNFTNotExists, typ)
	}

	burnedCoins := sdk.NewCoins(sdk.NewCoin(types.CreateDenom(typ, id), sdk.OneInt()))
	k.bk.SendCoinsFromAccountToModule(ctx, destroyer, types.ModuleName, burnedCoins)
	k.bk.BurnCoins(ctx, types.ModuleName, burnedCoins)

	// TODO Delete bank.Metadata (keeper method not available)

	nftStore := k.getNFTStore(ctx, typ)
	nftStore.Delete(types.GetNFTIdKey(id))
	return nil
}

func (k Keeper) GetNFT(ctx sdk.Context, typ, id string) (types.NFT, bool) {
	store := k.getNFTStore(ctx, typ)
	bz := store.Get(types.GetNFTIdKey(id))
	if len(bz) == 0 {
		return types.NFT{}, false
	}
	var nft types.NFT
	k.cdc.MustUnmarshal(bz, &nft)
	return nft, true
}

func (k Keeper) HasNFT(ctx sdk.Context, typ, id string) bool {
	store := k.getNFTStore(ctx, typ)
	return store.Has(types.GetNFTIdKey(id))
}

func (k Keeper) getNFTStore(ctx sdk.Context, typ string) prefix.Store {
	store := ctx.KVStore(k.storeKey)
	return prefix.NewStore(store, types.GetNFTKey(typ))
}
