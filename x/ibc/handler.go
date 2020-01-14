package ibc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	client "github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
)

// NewHandler defines the IBC handler
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		// IBC client msgs
		case client.MsgCreateClient:
			return client.HandleMsgCreateClient(ctx, k.ClientKeeper, msg)

		case client.MsgUpdateClient:
			return &sdk.Result{}, nil

		// IBC connection  msgs
		case connection.MsgConnectionOpenInit:
			return connection.HandleMsgConnectionOpenInit(ctx, k.ConnectionKeeper, msg)

		case connection.MsgConnectionOpenTry:
			return connection.HandleMsgConnectionOpenTry(ctx, k.ConnectionKeeper, msg)

		case connection.MsgConnectionOpenAck:
			return connection.HandleMsgConnectionOpenAck(ctx, k.ConnectionKeeper, msg)

		case connection.MsgConnectionOpenConfirm:
			return connection.HandleMsgConnectionOpenConfirm(ctx, k.ConnectionKeeper, msg)

		// IBC channel msgs
		case channel.MsgChannelOpenInit:
			return channel.HandleMsgChannelOpenInit(ctx, k.ChannelKeeper, msg)

		case channel.MsgChannelOpenTry:
			return channel.HandleMsgChannelOpenTry(ctx, k.ChannelKeeper, msg)

		case channel.MsgChannelOpenAck:
			return channel.HandleMsgChannelOpenAck(ctx, k.ChannelKeeper, msg)

		case channel.MsgChannelOpenConfirm:
			return channel.HandleMsgChannelOpenConfirm(ctx, k.ChannelKeeper, msg)

		case channel.MsgChannelCloseInit:
			return channel.HandleMsgChannelCloseInit(ctx, k.ChannelKeeper, msg)

		case channel.MsgChannelCloseConfirm:
			return channel.HandleMsgChannelCloseConfirm(ctx, k.ChannelKeeper, msg)

		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized IBC message type: %T", msg)
		}
	}
}
