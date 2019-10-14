package ics04

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/keeper"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
)

// NewHandler creates a new Handler instance for IBC connection
// transactions
func NewHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case types.MsgChannelOpenInit:
			return handleMsgChannelOpenInit(ctx, k, msg)

		case types.MsgChannelOpenTry:
			return handleMsgChannelOpenTry(ctx, k, msg)

		case types.MsgChannelOpenAck:
			return handleMsgChannelOpenAck(ctx, k, msg)

		case types.MsgChannelOpenConfirm:
			return handleMsgChannelOpenConfirm(ctx, k, msg)

		case types.MsgChannelCloseInit:
			return handleMsgChannelCloseInit(ctx, k, msg)

		case types.MsgChannelCloseConfirm:
			return handleMsgChannelCloseConfirm(ctx, k, msg)

		default:
			errMsg := fmt.Sprintf("unrecognized IBC connection message type: %T", msg)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleMsgChannelOpenInit(ctx sdk.Context, k keeper.Keeper, msg types.MsgChannelOpenInit) sdk.Result {
	_, err := k.ChanOpenInit(
		ctx, msg.Channel.Ordering, msg.Channel.ConnectionHops, msg.PortID, msg.ChannelID,
		msg.Channel.Counterparty, msg.Channel.Version,
	)
	if err != nil {
		return sdk.ResultFromError(err)
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeChannelOpenInit,
			sdk.NewAttribute(types.AttributeKeySenderPort, msg.PortID),
			sdk.NewAttribute(types.AttributeKeyChannelID, msg.ChannelID),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Signer.String()),
		),
	})

	return sdk.Result{Events: ctx.EventManager().Events()}
}

func handleMsgChannelOpenTry(ctx sdk.Context, k keeper.Keeper, msg types.MsgChannelOpenTry) sdk.Result {
	_, err := k.ChanOpenTry(ctx, msg.Channel.Ordering, msg.Channel.ConnectionHops, msg.PortID, msg.ChannelID,
		msg.Channel.Counterparty, msg.Channel.Version, msg.CounterpartyVersion, msg.ProofInit, msg.ProofHeight,
	)
	if err != nil {
		return sdk.ResultFromError(err)
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeChannelOpenTry,
			sdk.NewAttribute(types.AttributeKeyChannelID, msg.ChannelID),
			sdk.NewAttribute(types.AttributeKeySenderPort, msg.PortID), // TODO: double check sender and receiver
			sdk.NewAttribute(types.AttributeKeyReceiverPort, msg.Channel.Counterparty.PortID),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Signer.String()),
		),
	})

	return sdk.Result{Events: ctx.EventManager().Events()}
}

func handleMsgChannelOpenAck(ctx sdk.Context, k keeper.Keeper, msg types.MsgChannelOpenAck) sdk.Result {
	err := k.ChanOpenAck(
		ctx, msg.PortID, msg.ChannelID, msg.CounterpartyVersion, msg.ProofTry, msg.ProofHeight,
	)
	if err != nil {
		return sdk.ResultFromError(err)
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeChannelOpenAck,
			sdk.NewAttribute(types.AttributeKeySenderPort, msg.PortID),
			sdk.NewAttribute(types.AttributeKeyChannelID, msg.ChannelID),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Signer.String()),
		),
	})

	return sdk.Result{Events: ctx.EventManager().Events()}
}

func handleMsgChannelOpenConfirm(ctx sdk.Context, k keeper.Keeper, msg types.MsgChannelOpenConfirm) sdk.Result {
	err := k.ChanOpenConfirm(ctx, msg.PortID, msg.ChannelID, msg.ProofAck, msg.ProofHeight)
	if err != nil {
		return sdk.ResultFromError(err)
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeChannelOpenConfirm,
			sdk.NewAttribute(types.AttributeKeySenderPort, msg.PortID),
			sdk.NewAttribute(types.AttributeKeyChannelID, msg.ChannelID),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Signer.String()),
		),
	})

	return sdk.Result{Events: ctx.EventManager().Events()}
}

func handleMsgChannelCloseInit(ctx sdk.Context, k keeper.Keeper, msg types.MsgChannelCloseInit) sdk.Result {
	err := k.ChanCloseInit(ctx, msg.PortID, msg.ChannelID)
	if err != nil {
		return sdk.ResultFromError(err)
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeChannelCloseInit,
			sdk.NewAttribute(types.AttributeKeySenderPort, msg.PortID),
			sdk.NewAttribute(types.AttributeKeyChannelID, msg.ChannelID),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Signer.String()),
		),
	})

	return sdk.Result{Events: ctx.EventManager().Events()}
}

func handleMsgChannelCloseConfirm(ctx sdk.Context, k keeper.Keeper, msg types.MsgChannelCloseConfirm) sdk.Result {
	err := k.ChanCloseConfirm(ctx, msg.PortID, msg.ChannelID, msg.ProofInit, msg.ProofHeight)
	if err != nil {
		return sdk.ResultFromError(err)
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeChannelCloseConfirm,
			sdk.NewAttribute(types.AttributeKeySenderPort, msg.PortID),
			sdk.NewAttribute(types.AttributeKeyChannelID, msg.ChannelID),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Signer.String()),
		),
	})

	return sdk.Result{Events: ctx.EventManager().Events()}
}
