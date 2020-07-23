package channel

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/keeper"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
)

// HandleMsgChannelOpenInit defines the sdk.Handler for MsgChannelOpenInit
func HandleMsgChannelOpenInit(ctx sdk.Context, k keeper.Keeper, portCap *capabilitytypes.Capability, msg *types.MsgChannelOpenInit) (*sdk.Result, *capabilitytypes.Capability, error) {
	capKey, err := k.ChanOpenInit(
		ctx, msg.Channel.Ordering, msg.Channel.ConnectionHops, msg.PortID, msg.ChannelID,
		portCap, msg.Channel.Counterparty, msg.Channel.Version,
	)
	if err != nil {
		return nil, nil, sdkerrors.Wrap(err, "channel handshake open init failed")
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeChannelOpenInit,
			sdk.NewAttribute(types.AttributeKeyPortID, msg.PortID),
			sdk.NewAttribute(types.AttributeKeyChannelID, msg.ChannelID),
			sdk.NewAttribute(types.AttributeCounterpartyPortID, msg.Channel.Counterparty.PortID),
			sdk.NewAttribute(types.AttributeCounterpartyChannelID, msg.Channel.Counterparty.ChannelID),
			sdk.NewAttribute(types.AttributeKeyConnectionID, msg.Channel.ConnectionHops[0]),
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
func HandleMsgChannelOpenTry(ctx sdk.Context, k keeper.Keeper, portCap *capabilitytypes.Capability, msg *types.MsgChannelOpenTry) (*sdk.Result, *capabilitytypes.Capability, error) {
	// For now, convert uint64 heights to clientexported.Height
	proofHeight := clientexported.NewHeight(msg.ProofEpoch, msg.ProofHeight)
	capKey, err := k.ChanOpenTry(ctx, msg.Channel.Ordering, msg.Channel.ConnectionHops, msg.PortID, msg.ChannelID,
		portCap, msg.Channel.Counterparty, msg.Channel.Version, msg.CounterpartyVersion, msg.ProofInit, proofHeight,
	)
	if err != nil {
		return nil, nil, sdkerrors.Wrap(err, "channel handshake open try failed")
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeChannelOpenTry,
			sdk.NewAttribute(types.AttributeKeyPortID, msg.PortID),
			sdk.NewAttribute(types.AttributeKeyChannelID, msg.ChannelID),
			sdk.NewAttribute(types.AttributeCounterpartyPortID, msg.Channel.Counterparty.PortID),
			sdk.NewAttribute(types.AttributeCounterpartyChannelID, msg.Channel.Counterparty.ChannelID),
			sdk.NewAttribute(types.AttributeKeyConnectionID, msg.Channel.ConnectionHops[0]),
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
func HandleMsgChannelOpenAck(ctx sdk.Context, k keeper.Keeper, channelCap *capabilitytypes.Capability, msg *types.MsgChannelOpenAck) (*sdk.Result, error) {
	// For now, convert uint64 heights to clientexported.Height
	proofHeight := clientexported.NewHeight(msg.ProofEpoch, msg.ProofHeight)
	err := k.ChanOpenAck(
		ctx, msg.PortID, msg.ChannelID, channelCap, msg.CounterpartyVersion, msg.ProofTry, proofHeight,
	)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "channel handshake open ack failed")
	}

	channel, _ := k.GetChannel(ctx, msg.PortID, msg.ChannelID)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeChannelOpenAck,
			sdk.NewAttribute(types.AttributeKeyPortID, msg.PortID),
			sdk.NewAttribute(types.AttributeKeyChannelID, msg.ChannelID),
			sdk.NewAttribute(types.AttributeCounterpartyPortID, channel.Counterparty.PortID),
			sdk.NewAttribute(types.AttributeCounterpartyChannelID, channel.Counterparty.ChannelID),
			sdk.NewAttribute(types.AttributeKeyConnectionID, channel.ConnectionHops[0]),
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
func HandleMsgChannelOpenConfirm(ctx sdk.Context, k keeper.Keeper, channelCap *capabilitytypes.Capability, msg *types.MsgChannelOpenConfirm) (*sdk.Result, error) {
	// For now, convert uint64 heights to clientexported.Height
	proofHeight := clientexported.NewHeight(msg.ProofEpoch, msg.ProofHeight)
	err := k.ChanOpenConfirm(ctx, msg.PortID, msg.ChannelID, channelCap, msg.ProofAck, proofHeight)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "channel handshake open confirm failed")
	}

	channel, _ := k.GetChannel(ctx, msg.PortID, msg.ChannelID)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeChannelOpenConfirm,
			sdk.NewAttribute(types.AttributeKeyPortID, msg.PortID),
			sdk.NewAttribute(types.AttributeKeyChannelID, msg.ChannelID),
			sdk.NewAttribute(types.AttributeCounterpartyPortID, channel.Counterparty.PortID),
			sdk.NewAttribute(types.AttributeCounterpartyChannelID, channel.Counterparty.ChannelID),
			sdk.NewAttribute(types.AttributeKeyConnectionID, channel.ConnectionHops[0]),
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
func HandleMsgChannelCloseInit(ctx sdk.Context, k keeper.Keeper, channelCap *capabilitytypes.Capability, msg *types.MsgChannelCloseInit) (*sdk.Result, error) {
	err := k.ChanCloseInit(ctx, msg.PortID, msg.ChannelID, channelCap)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "channel handshake close init failed")
	}

	channel, _ := k.GetChannel(ctx, msg.PortID, msg.ChannelID)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeChannelCloseInit,
			sdk.NewAttribute(types.AttributeKeyPortID, msg.PortID),
			sdk.NewAttribute(types.AttributeKeyChannelID, msg.ChannelID),
			sdk.NewAttribute(types.AttributeCounterpartyPortID, channel.Counterparty.PortID),
			sdk.NewAttribute(types.AttributeCounterpartyChannelID, channel.Counterparty.ChannelID),
			sdk.NewAttribute(types.AttributeKeyConnectionID, channel.ConnectionHops[0]),
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
func HandleMsgChannelCloseConfirm(ctx sdk.Context, k keeper.Keeper, channelCap *capabilitytypes.Capability, msg *types.MsgChannelCloseConfirm) (*sdk.Result, error) {
	// For now, convert uint64 heights to clientexported.Height
	proofHeight := clientexported.NewHeight(msg.ProofEpoch, msg.ProofHeight)
	err := k.ChanCloseConfirm(ctx, msg.PortID, msg.ChannelID, channelCap, msg.ProofInit, proofHeight)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "channel handshake close confirm failed")
	}

	channel, _ := k.GetChannel(ctx, msg.PortID, msg.ChannelID)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeChannelCloseConfirm,
			sdk.NewAttribute(types.AttributeKeyPortID, msg.PortID),
			sdk.NewAttribute(types.AttributeKeyChannelID, msg.ChannelID),
			sdk.NewAttribute(types.AttributeCounterpartyPortID, channel.Counterparty.PortID),
			sdk.NewAttribute(types.AttributeCounterpartyChannelID, channel.Counterparty.ChannelID),
			sdk.NewAttribute(types.AttributeKeyConnectionID, channel.ConnectionHops[0]),
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
