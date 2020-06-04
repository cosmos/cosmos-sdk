package channel

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/capability"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/keeper"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
)

// HandleMsgChannelOpenInit defines the sdk.Handler for MsgChannelOpenInit
func HandleMsgChannelOpenInit(ctx sdk.Context, k keeper.Keeper, portCap *capability.Capability, msg *MsgChannelOpenInit) (*sdk.Result, *capability.Capability, error) {
	capKey, err := k.ChanOpenInit(
		ctx, msg.Channel.Ordering, msg.Channel.ConnectionHops, msg.PortID, msg.ChannelID,
		portCap, msg.Channel.Counterparty, msg.Channel.Version,
	)
	if err != nil {
		return nil, nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeChannelOpenInit,
			sdk.NewAttribute(types.AttributeKeyPortID, msg.PortID),
			sdk.NewAttribute(types.AttributeKeyChannelID, msg.ChannelID),
			sdk.NewAttribute(types.AttributeCounterpartyPortID, msg.Channel.Counterparty.PortID),
			sdk.NewAttribute(types.AttributeCounterpartyChannelID, msg.Channel.Counterparty.ChannelID),
			sdk.NewAttribute(types.AttributeKeyConnectionID, msg.Channel.ConnectionHops[0]), // TODO: iterate
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	})

	return &sdk.Result{
		Events: ctx.EventManager().Events().ToABCIEvents(),
	}, capKey, nil
}

// HandleMsgChannelOpenTry defines the sdk.Handler for MsgChannelOpenTry
func HandleMsgChannelOpenTry(ctx sdk.Context, k keeper.Keeper, portCap *capability.Capability, msg *MsgChannelOpenTry) (*sdk.Result, *capability.Capability, error) {
	capKey, err := k.ChanOpenTry(ctx, msg.Channel.Ordering, msg.Channel.ConnectionHops, msg.PortID, msg.ChannelID,
		portCap, msg.Channel.Counterparty, msg.Channel.Version, msg.CounterpartyVersion, msg.ProofInit, msg.ProofHeight,
	)
	if err != nil {
		return nil, nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeChannelOpenTry,
			sdk.NewAttribute(types.AttributeKeyPortID, msg.PortID),
			sdk.NewAttribute(types.AttributeKeyChannelID, msg.ChannelID),
			sdk.NewAttribute(types.AttributeCounterpartyPortID, msg.Channel.Counterparty.PortID),
			sdk.NewAttribute(types.AttributeCounterpartyChannelID, msg.Channel.Counterparty.ChannelID),
			sdk.NewAttribute(types.AttributeKeyConnectionID, msg.Channel.ConnectionHops[0]), // TODO: iterate
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	})

	return &sdk.Result{
		Events: ctx.EventManager().Events().ToABCIEvents(),
	}, capKey, nil
}

// HandleMsgChannelOpenAck defines the sdk.Handler for MsgChannelOpenAck
func HandleMsgChannelOpenAck(ctx sdk.Context, k keeper.Keeper, channelCap *capability.Capability, msg *MsgChannelOpenAck) (*sdk.Result, error) {
	err := k.ChanOpenAck(
		ctx, msg.PortID, msg.ChannelID, channelCap, msg.CounterpartyVersion, msg.ProofTry, msg.ProofHeight,
	)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeChannelOpenAck,
			sdk.NewAttribute(types.AttributeKeyPortID, msg.PortID),
			sdk.NewAttribute(types.AttributeKeyChannelID, msg.ChannelID),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	})

	return &sdk.Result{
		Events: ctx.EventManager().Events().ToABCIEvents(),
	}, nil
}

// HandleMsgChannelOpenConfirm defines the sdk.Handler for MsgChannelOpenConfirm
func HandleMsgChannelOpenConfirm(ctx sdk.Context, k keeper.Keeper, channelCap *capability.Capability, msg *MsgChannelOpenConfirm) (*sdk.Result, error) {
	err := k.ChanOpenConfirm(ctx, msg.PortID, msg.ChannelID, channelCap, msg.ProofAck, msg.ProofHeight)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeChannelOpenConfirm,
			sdk.NewAttribute(types.AttributeKeyPortID, msg.PortID),
			sdk.NewAttribute(types.AttributeKeyChannelID, msg.ChannelID),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	})

	return &sdk.Result{
		Events: ctx.EventManager().Events().ToABCIEvents(),
	}, nil
}

// HandleMsgChannelCloseInit defines the sdk.Handler for MsgChannelCloseInit
func HandleMsgChannelCloseInit(ctx sdk.Context, k keeper.Keeper, channelCap *capability.Capability, msg *MsgChannelCloseInit) (*sdk.Result, error) {
	err := k.ChanCloseInit(ctx, msg.PortID, msg.ChannelID, channelCap)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeChannelCloseInit,
			sdk.NewAttribute(types.AttributeKeyPortID, msg.PortID),
			sdk.NewAttribute(types.AttributeKeyChannelID, msg.ChannelID),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	})

	return &sdk.Result{
		Events: ctx.EventManager().Events().ToABCIEvents(),
	}, nil
}

// HandleMsgChannelCloseConfirm defines the sdk.Handler for MsgChannelCloseConfirm
func HandleMsgChannelCloseConfirm(ctx sdk.Context, k keeper.Keeper, channelCap *capability.Capability, msg *MsgChannelCloseConfirm) (*sdk.Result, error) {
	err := k.ChanCloseConfirm(ctx, msg.PortID, msg.ChannelID, channelCap, msg.ProofInit, msg.ProofHeight)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeChannelCloseConfirm,
			sdk.NewAttribute(types.AttributeKeyPortID, msg.PortID),
			sdk.NewAttribute(types.AttributeKeyChannelID, msg.ChannelID),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	})

	return &sdk.Result{
		Events: ctx.EventManager().Events().ToABCIEvents(),
	}, nil
}
