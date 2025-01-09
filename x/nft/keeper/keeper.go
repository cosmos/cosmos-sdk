package keeper

import (
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/x/nft"

	"github.com/cosmos/cosmos-sdk/codec"
)

// Keeper of the nft store
type Keeper struct {
	appmodule.Environment

	cdc codec.BinaryCodec
	bk  nft.BankKeeper
	ac  address.Codec
}

// NewKeeper creates a new nft Keeper instance
func NewKeeper(env appmodule.Environment,
	cdc codec.BinaryCodec, ak nft.AccountKeeper, bk nft.BankKeeper,
) Keeper {
	// ensure nft module account is set
	if addr := ak.GetModuleAddress(nft.ModuleName); addr == nil {
		panic("the nft module account has not been set")
	}

	return Keeper{
		Environment: env,
		cdc:         cdc,
		bk:          bk,
		ac:          ak.AddressCodec(),
	}
}

// MsgNewClass creates a new NFT class
func (k Keeper) MsgNewClass(ctx context.Context, msg *nft.MsgNewClass) (*nft.MsgNewClassResponse, error) {
	class := nft.Class{
		Id:          msg.ClassId,
		Name:        msg.Name,
		Symbol:      msg.Symbol,
		Description: msg.Description,
		Uri:         msg.Uri,
		UriHash:     msg.UriHash,
		Data:        msg.Data,
	}
	if err := k.SaveClass(ctx, class); err != nil {
		return nil, err
	}
	return &nft.MsgNewClassResponse{}, nil
}

// MsgUpdateClass updates an existing NFT class
func (k Keeper) MsgUpdateClass(ctx context.Context, msg *nft.MsgUpdateClass) (*nft.MsgUpdateClassResponse, error) {
	class := nft.Class{
		Id:          msg.ClassId,
		Name:        msg.Name,
		Symbol:      msg.Symbol,
		Description: msg.Description,
		Uri:         msg.Uri,
		UriHash:     msg.UriHash,
		Data:        msg.Data,
	}
	if err := k.UpdateClass(ctx, class); err != nil {
		return nil, err
	}
	return &nft.MsgUpdateClassResponse{}, nil
}

// MsgMintNFT mints a new NFT
func (k Keeper) MsgMintNFT(ctx context.Context, msg *nft.MsgMintNFT) (*nft.MsgMintNFTResponse, error) {
	nft := nft.NFT{
		ClassId: msg.ClassId,
		Id:      msg.Id,
		Uri:     msg.Uri,
		UriHash: msg.UriHash,
		Data:    msg.Data,
	}
	if err := k.Mint(ctx, nft, msg.Receiver); err != nil {
		return nil, err
	}
	return &nft.MsgMintNFTResponse{}, nil
}

// MsgBurnNFT burns an existing NFT
func (k Keeper) MsgBurnNFT(ctx context.Context, msg *nft.MsgBurnNFT) (*nft.MsgBurnNFTResponse, error) {
	if err := k.Burn(ctx, msg.ClassId, msg.Id); err != nil {
		return nil, err
	}
	return &nft.MsgBurnNFTResponse{}, nil
}

// MsgUpdateNFT updates an existing NFT
func (k Keeper) MsgUpdateNFT(ctx context.Context, msg *nft.MsgUpdateNFT) (*nft.MsgUpdateNFTResponse, error) {
	nft := nft.NFT{
		ClassId: msg.ClassId,
		Id:      msg.Id,
		Uri:     msg.Uri,
		UriHash: msg.UriHash,
		Data:    msg.Data,
	}
	if err := k.Update(ctx, nft); err != nil {
		return nil, err
	}
	return &nft.MsgUpdateNFTResponse{}, nil
}
