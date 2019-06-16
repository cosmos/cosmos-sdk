package ibc

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		fmt.Println(msg)
		switch msg := msg.(type) {
		case MsgCreateClient:
			return handleMsgCreateClient(ctx, msg, k)
		case MsgOpenConnection:
			return handleMsgOpenConnection(ctx, msg, k)
		case MsgUpdateClient:
			return handleMsgUpdateClient(ctx, msg, k)
		case MsgOpenChannel:
			return handleMsgOpenChannel(ctx, msg, k)
		default:
			return sdk.ErrUnknownRequest("unrecognized IBC message tyoe").Result()
		}
	}
}

func handleMsgCreateClient(ctx sdk.Context, msg MsgCreateClient, k Keeper) sdk.Result {
	err := k.CreateClient(ctx, msg.ClientID, msg.ConsensusState)
	if err != nil {
		return sdk.NewError(sdk.CodespaceType("ibc"), sdk.CodeType(100), err.Error()).Result()
	}
	return sdk.Result{}
}

func handleMsgUpdateClient(ctx sdk.Context, msg MsgUpdateClient, k Keeper) sdk.Result {
	err := k.UpdateClient(ctx, msg.ClientID, msg.Header)
	if err != nil {
		return sdk.NewError(sdk.CodespaceType("ibc"), sdk.CodeType(300), err.Error()).Result()
	}
	return sdk.Result{}
}

func handleMsgOpenConnection(ctx sdk.Context, msg MsgOpenConnection, k Keeper) sdk.Result {
	err := k.OpenConnection(ctx, msg.ConnectionID, msg.CounterpartyID, msg.ClientID, msg.CounterpartyClientID)
	if err != nil {
		fmt.Println(222, err)
		return sdk.NewError(sdk.CodespaceType("ibc"), sdk.CodeType(200), err.Error()).Result()
	}
	return sdk.Result{}
}

func handleMsgOpenChannel(ctx sdk.Context, msg MsgOpenChannel, k Keeper) sdk.Result {
	err := k.OpenChannel(ctx, msg.ModuleID, msg.ConnectionID, msg.ChannelID, msg.CounterpartyID, msg.CounterpartyModuleID)
	if err != nil {
		return sdk.NewError(sdk.CodespaceType("ibc"), sdk.CodeType(400), err.Error()).Result()
	}
	return sdk.Result{}
}
