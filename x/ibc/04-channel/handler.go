package channel

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/keeper"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
)

// HandleMsgChannelOpenInit defines the sdk.Handler for MsgChannelOpenInit
func HandleMsgChannelOpenInit(ctx sdk.Context, k keeper.Keeper, msg types.MsgChannelOpenInit) sdk.Result {
	err := k.ChanOpenInit(
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

// HandleMsgChannelOpenTry defines the sdk.Handler for MsgChannelOpenTry
func HandleMsgChannelOpenTry(ctx sdk.Context, k keeper.Keeper, msg types.MsgChannelOpenTry) sdk.Result {
	err := k.ChanOpenTry(ctx, msg.Channel.Ordering, msg.Channel.ConnectionHops, msg.PortID, msg.ChannelID,
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

// HandleMsgChannelOpenAck defines the sdk.Handler for MsgChannelOpenAck
func HandleMsgChannelOpenAck(ctx sdk.Context, k keeper.Keeper, msg types.MsgChannelOpenAck) sdk.Result {
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

// HandleMsgChannelOpenConfirm defines the sdk.Handler for MsgChannelOpenConfirm
func HandleMsgChannelOpenConfirm(ctx sdk.Context, k keeper.Keeper, msg types.MsgChannelOpenConfirm) sdk.Result {
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

// HandleMsgChannelCloseInit defines the sdk.Handler for MsgChannelCloseInit
func HandleMsgChannelCloseInit(ctx sdk.Context, k keeper.Keeper, msg types.MsgChannelCloseInit) sdk.Result {
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

// HandleMsgChannelCloseConfirm defines the sdk.Handler for MsgChannelCloseConfirm
func HandleMsgChannelCloseConfirm(ctx sdk.Context, k keeper.Keeper, msg types.MsgChannelCloseConfirm) sdk.Result {
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
