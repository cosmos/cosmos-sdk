package client

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func HandleMsgCreateClient(ctx sdk.Context, msg MsgCreateClient, man Manager) sdk.Result {
	_, err := man.Create(ctx, msg.ClientID, msg.ConsensusState)
	if err != nil {
		return sdk.NewError(sdk.CodespaceType("ibc"), sdk.CodeType(100), err.Error()).Result()
	}

	// TODO: events
	return sdk.Result{}
}

func HandleMsgUpdateClient(ctx sdk.Context, msg MsgUpdateClient, man Manager) sdk.Result {
	obj, err := man.Query(ctx, msg.ClientID)
	if err != nil {
		return sdk.NewError(sdk.CodespaceType("ibc"), sdk.CodeType(200), err.Error()).Result()
	}
	err = obj.Update(ctx, msg.Header)
	if err != nil {
		return sdk.NewError(sdk.CodespaceType("ibc"), sdk.CodeType(300), err.Error()).Result()
	}

	// TODO: events
	return sdk.Result{}
}
