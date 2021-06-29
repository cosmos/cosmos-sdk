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
	if k.HasType(ctx, typ) {
		return sdkerrors.Wrap(types.ErrNFTTypeExists, typ)
	}

	if k.HasNFT(ctx, typ, id) {
		return sdkerrors.Wrap(types.ErrNFTTypeExists, id)
	}

	metadata := k.GetMetadata(ctx, typ)
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
	return nil
}

// SendNFT defines a method for sending a nft from one account to another account.
func (k Keeper) SendNFT(ctx sdk.Context,
	typ string,
	id string,
	sender sdk.AccAddress,
	receiver sdk.AccAddress) error {
	return nil
}

// BurnNFT defines a method for burning a nft from a specific account.
func (k Keeper) BurnNFT(ctx sdk.Context,
	typ string,
	id string,
	destroyer sdk.AccAddress) error {
	return nil
}

func (k Keeper) HasNFT(ctx sdk.Context, typ, id string) bool {
	store := k.getNFTStore(ctx, typ)
	return store.Has(types.GetNFTIdKey(id))
}

func (k Keeper) getNFTStore(ctx sdk.Context, typ string) prefix.Store {
	store := ctx.KVStore(k.storeKey)
	return prefix.NewStore(store, types.GetNFTKey(typ))
}
