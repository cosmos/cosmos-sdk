package ics02

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/keeper"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
)

// NewHandler creates a new Handler instance for IBC client
// transactions
func NewHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case types.MsgCreateClient:
			return handleMsgCreateClient(ctx, k, msg)

		case types.MsgUpdateClient:
			return handleMsgUpdateClient(ctx, k, msg)

		case types.MsgSubmitMisbehaviour:
			return handleMsgSubmitMisbehaviour(ctx, k, msg)

		default:
			errMsg := fmt.Sprintf("unrecognized IBC Client message type: %T", msg)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleMsgCreateClient(ctx sdk.Context, k keeper.Keeper, msg types.MsgCreateClient) sdk.Result {
	_, err := k.CreateClient(ctx, msg.ClientID, msg.ConsensusState)
	if err != nil {
		return sdk.ResultFromError(err)
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeCreateClient,
			sdk.NewAttribute(types.AttributeKeyClientID, msg.ClientID),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Signer.String()),
		),
	})

	return sdk.Result{Events: ctx.EventManager().Events()}
}

func handleMsgUpdateClient(ctx sdk.Context, k keeper.Keeper, msg types.MsgUpdateClient) sdk.Result {
	state, err := k.Query(ctx, msg.ClientID)
	if err != nil {
		return sdk.ResultFromError(err)
	}

	err = k.Update(ctx, state, msg.Header)
	if err != nil {
		return sdk.ResultFromError(err)
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeUpdateClient,
			sdk.NewAttribute(types.AttributeKeyClientID, msg.ClientID),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Signer.String()),
		),
	})

	// TODO: events
	return sdk.Result{Events: ctx.EventManager().Events()}
}

func handleMsgSubmitMisbehaviour(ctx sdk.Context, k keeper.Keeper, msg types.MsgSubmitMisbehaviour) sdk.Result {
	state, err := k.Query(ctx, msg.ClientID)
	if err != nil {
		return sdk.ResultFromError(err)
	}

	err = k.CheckMisbehaviourAndUpdateState(ctx, state, msg.Evidence)
	if err != nil {
		return sdk.ResultFromError(err)
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeSubmitMisbehaviour,
			sdk.NewAttribute(types.AttributeKeyClientID, msg.ClientID),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Signer.String()),
		),
	})

	return sdk.Result{Events: ctx.EventManager().Events()}
}
