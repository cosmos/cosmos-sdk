package ibc

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	client "github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	transfer "github.com/cosmos/cosmos-sdk/x/ibc/20-transfer"
)

// NewHandler defines the IBC handler
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		// IBC client msgs
		case client.MsgCreateClient:
			return client.HandleMsgCreateClient(ctx, k.ClientKeeper, msg)

		case client.MsgUpdateClient:
			return client.HandleMsgUpdateClient(ctx, k.ClientKeeper, msg)

		case client.MsgSubmitMisbehaviour:
			return client.HandleMsgSubmitMisbehaviour(ctx, k.ClientKeeper, msg)

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

		// IBC transfer msgs
		case transfer.MsgTransfer:
			return transfer.HandleMsgTransfer(ctx, k.TransferKeeper, msg)

		case transfer.MsgRecvPacket:
			return transfer.HandleMsgRecvPacket(ctx, k.TransferKeeper, msg)

		default:
			errMsg := fmt.Sprintf("unrecognized IBC message type: %T", msg)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}
