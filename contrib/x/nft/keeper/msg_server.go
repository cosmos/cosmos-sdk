package keeper

import (
	"bytes"
	"context"

	"google.golang.org/grpc"

	errorsmod "cosmossdk.io/errors"

	nft "github.com/cosmos/cosmos-sdk/contrib/x/nft"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type msgServer struct {
	Keeper
}

var _ nft.MsgServer = msgServer{}

// newMsgServerImpl returns an implementation of the x/nft MsgServer
// interface for the provided Keeper.
func newMsgServerImpl(k Keeper) nft.MsgServer {
	return msgServer{Keeper: k}
}

// RegisterMsgServer registers the x/nft Msg service on r. The msg server
// is unexported: code outside the module runs a msg through the
// MsgServiceRouter with msgservice.Dispatch.
func RegisterMsgServer(r grpc.ServiceRegistrar, k Keeper) {
	nft.RegisterMsgServer(r, newMsgServerImpl(k))
}

// Send implements Send method of the types.MsgServer.
func (k msgServer) Send(goCtx context.Context, msg *nft.MsgSend) (*nft.MsgSendResponse, error) {
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

	ctx := sdk.UnwrapSDKContext(goCtx)
	owner := k.GetOwner(ctx, msg.ClassId, msg.Id)
	if !bytes.Equal(owner, sender) {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "%s is not the owner of nft %s", msg.Sender, msg.Id)
	}

	if err := k.Transfer(ctx, msg.ClassId, msg.Id, receiver); err != nil {
		return nil, err
	}

	err = ctx.EventManager().EmitTypedEvent(&nft.EventSend{
		ClassId:  msg.ClassId,
		Id:       msg.Id,
		Sender:   msg.Sender,
		Receiver: msg.Receiver,
	})
	return &nft.MsgSendResponse{}, err
}
