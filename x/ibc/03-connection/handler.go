package connection

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/keeper"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
)

// NewHandler creates a new Handler instance for IBC connection
// transactions
func NewHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case types.MsgConnectionOpenInit:
			return handleMsgConnectionOpenInit(ctx, k, msg)

		case types.MsgConnectionOpenTry:
			return handleMsgConnectionOpenTry(ctx, k, msg)

		case types.MsgConnectionOpenAck:
			return handleMsgConnectionOpenAck(ctx, k, msg)

		case types.MsgConnectionOpenConfirm:
			return handleMsgConnectionOpenConfirm(ctx, k, msg)

		default:
			errMsg := fmt.Sprintf("unrecognized IBC connection message type: %T", msg)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleMsgConnectionOpenInit(ctx sdk.Context, k keeper.Keeper, msg types.MsgConnectionOpenInit) sdk.Result {
	_, err := k.ConnOpenInit(ctx, msg.ConnectionID, msg.ClientID, msg.Counterparty)
	if err != nil {
		return sdk.ResultFromError(err)
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeConnectionOpenInit,
			sdk.NewAttribute(types.AttributeKeyConnectionID, msg.ConnectionID),
			sdk.NewAttribute(types.AttributeKeyCounterpartyClientID, msg.Counterparty.ClientID),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Signer.String()),
		),
	})

	return sdk.Result{Events: ctx.EventManager().Events()}
}

func handleMsgConnectionOpenTry(ctx sdk.Context, k keeper.Keeper, msg types.MsgConnectionOpenTry) sdk.Result {
	_, err := k.ConnOpenTry(ctx, msg.ConnectionID, msg.ClientID, msg.Counterparty, msg.Proofs, msg.Height)
	if err != nil {
		return sdk.ResultFromError(err)
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeConnectionOpenTry,
			sdk.NewAttribute(types.AttributeKeyConnectionID, msg.ConnectionID),
			sdk.NewAttribute(types.AttributeKeyCounterpartyClientID, msg.Counterparty.ClientID),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Signer.String()),
		),
	})

	return sdk.Result{Events: ctx.EventManager().Events()}
}

func handleMsgConnectionOpenAck(ctx sdk.Context, k keeper.Keeper, msg types.MsgConnectionOpenAck) sdk.Result {
	_, err := k.OpenAck(ctx, msg.Proofs, msg.Height, msg.ConnectionID)
	if err != nil {
		return sdk.ResultFromError(err)
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeConnectionOpenAck,
			sdk.NewAttribute(types.AttributeKeyConnectionID, msg.ConnectionID),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Signer.String()),
		),
	})

	return sdk.Result{Events: ctx.EventManager().Events()}
}

func handleMsgConnectionOpenConfirm(ctx sdk.Context, k keeper.Keeper, msg types.MsgConnectionOpenConfirm) sdk.Result {
	_, err := k.OpenConfirm(ctx, msg.Proofs, msg.Height, msg.ConnectionID)
	if err != nil {
		return sdk.ResultFromError(err)
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeConnectionOpenConfirm,
			sdk.NewAttribute(types.AttributeKeyConnectionID, msg.ConnectionID),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Signer.String()),
		),
	})

	return sdk.Result{Events: ctx.EventManager().Events()}
}
