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

	if err = k.env.EventService.EventManager(ctx).Emit(&nft.EventSend{
		ClassId:  msg.ClassId,
		Id:       msg.Id,
		Sender:   msg.Sender,
		Receiver: msg.Receiver,
	}); err != nil {
		return nil, err
	}

	return &nft.MsgSendResponse{}, nil
}
