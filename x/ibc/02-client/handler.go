package ics02

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewHandler creates a new Handler instance for IBC client
// transactions
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case MsgCreateClient:
			return handleMsgCreateClient(ctx, k, msg)

		case MsgUpdateClient:
			return handleMsgUpdateClient(ctx, k, msg)

		case MsgSubmitMisbehaviour:
			return handleMsgSubmitMisbehaviour(ctx, k, msg)

		default:
			errMsg := fmt.Sprintf("unrecognized IBC Client message type: %T", msg)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleMsgCreateClient(ctx sdk.Context, k Keeper, msg MsgCreateClient) sdk.Result {
	_, err := k.CreateClient(ctx, msg.ClientID, msg.ClientType, msg.ConsensusState)
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

func handleMsgUpdateClient(ctx sdk.Context, k Keeper, msg MsgUpdateClient) sdk.Result {
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

	// TODO: events
	return sdk.Result{Events: ctx.EventManager().Events()}
}

func handleMsgSubmitMisbehaviour(ctx sdk.Context, k Keeper, msg MsgSubmitMisbehaviour) sdk.Result {
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
