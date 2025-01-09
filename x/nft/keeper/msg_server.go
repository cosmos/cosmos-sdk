package keeper

import (
	"bytes"
	"context"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/x/nft"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ nft.MsgServer = Keeper{}

// Send implements Send method of the types.MsgServer.
func (k Keeper) Send(ctx context.Context, msg *nft.MsgSend) (*nft.MsgSendResponse, error) {
	if len(msg.ClassId) == 0 {
		return nil, nft.ErrEmptyClassID
	}

	if len(msg.Id) == 0 {
		return nil, nft.ErrEmptyNFTID
	}

	sender, err := k.ac.StringToBytes(msg.Sender)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", msg.Sender)
	}

	receiver, err := k.ac.StringToBytes(msg.Receiver)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid receiver address (%s)", msg.Receiver)
	}

	owner := k.GetOwner(ctx, msg.ClassId, msg.Id)
	if !bytes.Equal(owner, sender) {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "%s is not the owner of nft %s", msg.Sender, msg.Id)
	}

	if err := k.Transfer(ctx, msg.ClassId, msg.Id, receiver); err != nil {
		return nil, err
	}

	if err = k.EventService.EventManager(ctx).Emit(&nft.EventSend{
		ClassId:  msg.ClassId,
		Id:       msg.Id,
		Sender:   msg.Sender,
		Receiver: msg.Receiver,
	}); err != nil {
		return nil, err
	}

	return &nft.MsgSendResponse{}, nil
}

// NewClass implements NewClass method of the types.MsgServer.
func (k Keeper) NewClass(ctx context.Context, msg *nft.MsgNewClass) (*nft.MsgNewClassResponse, error) {
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

// UpdateClass implements UpdateClass method of the types.MsgServer.
func (k Keeper) UpdateClass(ctx context.Context, msg *nft.MsgUpdateClass) (*nft.MsgUpdateClassResponse, error) {
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

// MintNFT implements MintNFT method of the types.MsgServer.
func (k Keeper) MintNFT(ctx context.Context, msg *nft.MsgMintNFT) (*nft.MsgMintNFTResponse, error) {
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

// BurnNFT implements BurnNFT method of the types.MsgServer.
func (k Keeper) BurnNFT(ctx context.Context, msg *nft.MsgBurnNFT) (*nft.MsgBurnNFTResponse, error) {
	if err := k.Burn(ctx, msg.ClassId, msg.Id); err != nil {
		return nil, err
	}
	return &nft.MsgBurnNFTResponse{}, nil
}

// UpdateNFT implements UpdateNFT method of the types.MsgServer.
func (k Keeper) UpdateNFT(ctx context.Context, msg *nft.MsgUpdateNFT) (*nft.MsgUpdateNFTResponse, error) {
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
