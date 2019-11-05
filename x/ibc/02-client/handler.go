package client

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	exported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
)

// HandleMsgCreateClient defines the sdk.Handler for MsgCreateClient
func HandleMsgCreateClient(ctx sdk.Context, k Keeper, msg MsgCreateClient) sdk.Result {
	clientType, err := exported.ClientTypeFromString(msg.ClientType)
	if err != nil {
		return sdk.ResultFromError(ErrInvalidClientType(DefaultCodespace, err.Error()))
	}

	// TODO: should we create an event with the new client state id ?
	_, err = k.CreateClient(ctx, msg.ClientID, clientType, msg.ConsensusState)
	if err != nil {
		return sdk.ResultFromError(err)
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			EventTypeCreateClient,
			sdk.NewAttribute(AttributeKeyClientID, msg.ClientID),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Signer.String()),
		),
	})

	return sdk.Result{Events: ctx.EventManager().Events()}
}

// HandleMsgUpdateClient defines the sdk.Handler for MsgUpdateClient
func HandleMsgUpdateClient(ctx sdk.Context, k Keeper, msg MsgUpdateClient) sdk.Result {
	err := k.UpdateClient(ctx, msg.ClientID, msg.Header)
	if err != nil {
		return sdk.ResultFromError(err)
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			EventTypeUpdateClient,
			sdk.NewAttribute(AttributeKeyClientID, msg.ClientID),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Signer.String()),
		),
	})

	return sdk.Result{Events: ctx.EventManager().Events()}
}

// HandleMsgSubmitMisbehaviour defines the sdk.Handler for MsgSubmitMisbehaviour
func HandleMsgSubmitMisbehaviour(ctx sdk.Context, k Keeper, msg MsgSubmitMisbehaviour) sdk.Result {
	err := k.CheckMisbehaviourAndUpdateState(ctx, msg.ClientID, msg.Evidence)
	if err != nil {
		return sdk.ResultFromError(err)
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			EventTypeSubmitMisbehaviour,
			sdk.NewAttribute(AttributeKeyClientID, msg.ClientID),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Signer.String()),
		),
	})

	return sdk.Result{Events: ctx.EventManager().Events()}
}
