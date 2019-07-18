package channel

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func HandleMsgOpenInit(ctx sdk.Context, msg MsgOpenInit, man Handshaker) sdk.Result {
	_, err := man.OpenInit(ctx, msg.ConnectionID, msg.ChannelID, msg.Channel, msg.NextTimeout)
	if err != nil {
		return sdk.NewError(sdk.CodespaceType("ibc"), 100, "").Result()
	}
	return sdk.Result{}
}

func HandleMsgOpenTry(ctx sdk.Context, msg MsgOpenTry, man Handshaker) sdk.Result {
	_, err := man.OpenTry(ctx, msg.Proofs, msg.ConnectionID, msg.ChannelID, msg.Channel, msg.Timeout, msg.NextTimeout)
	if err != nil {
		return sdk.NewError(sdk.CodespaceType("ibc"), 200, "").Result()
	}
	return sdk.Result{}
}

func HandleMsgOpenAck(ctx sdk.Context, msg MsgOpenAck, man Handshaker) sdk.Result {
	_, err := man.OpenAck(ctx, msg.Proofs, msg.ConnectionID, msg.ChannelID, msg.Timeout, msg.NextTimeout)
	if err != nil {
		return sdk.NewError(sdk.CodespaceType("ibc"), 300, "").Result()
	}
	return sdk.Result{}
}

func HandleMsgOpenConfirm(ctx sdk.Context, msg MsgOpenConfirm, man Handshaker) sdk.Result {
	_, err := man.OpenConfirm(ctx, msg.Proofs, msg.ConnectionID, msg.ChannelID, msg.Timeout)
	if err != nil {
		return sdk.NewError(sdk.CodespaceType("ibc"), 400, "").Result()
	}
	return sdk.Result{}
}

type Handler func(sdk.Context, Packet) sdk.Result

func HandleMsgReceive(ctx sdk.Context, msg MsgReceive, man Manager) sdk.Result {
	err := man.Receive(ctx, msg.Proofs, msg.ConnectionID, msg.ChannelID, msg.Packet)
	if err != nil {
		return sdk.NewError(sdk.CodespaceType("ibc"), 500, "").Result()
	}
	handler := man.router.Route(msg.Packet.Route())
	return handler(ctx, msg.Packet)
}
