package ibc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgCreateClient:
			return handleMsgCreateClient(ctx, msg, k)
		case MsgOpenConnection:
			return handleMsgOpenConnection(ctx, msg, k)
		default:
			return sdk.ErrUnknownRequest("unrecognized IBC message tyoe").Result()
		}
	}
}

func handleMsgCreateClient(ctx sdk.Context, msg MsgCreateClient, k Keeper) sdk.Result {
	err := k.CreateClient(ctx, msg.ClientID, msg.ConsensusState)
	if err != nil {
		return sdk.NewError(sdk.CodespaceRoot, sdk.CodeType(100), err.Error()).Result()
	}
	return sdk.Result{}
}

func handleMsgOpenConnection(ctx sdk.Context, msg MsgOpenConnection, k Keeper) sdk.Result {
	err := k.OpenConnection(ctx, msg.ConnectionID, msg.CounterpartyID, msg.ClientID, msg.CounterpartyClientID)
	if err != nil {
		return sdk.NewError(sdk.CodespaceRoot, sdk.CodeType(200), err.Error()).Result()
	}
	return sdk.Result{}
}
