package connection

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func HandleMsgOpenInit(ctx sdk.Context, msg MsgOpenInit, man Handshaker) sdk.Result {
	_, err := man.OpenInit(ctx, msg.ConnectionID, msg.Connection, msg.CounterpartyClient)
	if err != nil {
		return sdk.NewError(sdk.CodespaceType("ibc"), 100, "").Result()
	}
	return sdk.Result{}
}

func HandleMsgOpenTry(ctx sdk.Context, msg MsgOpenTry, man Handshaker) sdk.Result {
	_, err := man.OpenTry(ctx, msg.Proofs, msg.ConnectionID, msg.Connection, msg.CounterpartyClient)
	if err != nil {
		return sdk.NewError(sdk.CodespaceType("ibc"), 200, "").Result()
	}
	return sdk.Result{}
}

func HandleMsgOpenAck(ctx sdk.Context, msg MsgOpenAck, man Handshaker) sdk.Result {
	_, err := man.OpenAck(ctx, msg.Proofs, msg.ConnectionID)
	if err != nil {
		return sdk.NewError(sdk.CodespaceType("ibc"), 300, "").Result()
	}
	return sdk.Result{}
}

func HandleMsgOpenConfirm(ctx sdk.Context, msg MsgOpenConfirm, man Handshaker) sdk.Result {
	_, err := man.OpenConfirm(ctx, msg.Proofs, msg.ConnectionID)
	if err != nil {
		return sdk.NewError(sdk.CodespaceType("ibc"), 400, "").Result()
	}
	return sdk.Result{}
}
