package channel

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/keeper"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
)

// HandleMsgChannelOpenInit defines the sdk.Handler for MsgChannelOpenInit
func HandleMsgChannelOpenInit(ctx sdk.Context, k keeper.Keeper, portCap *capabilitytypes.Capability, msg *types.MsgChannelOpenInit) (*sdk.Result, string, *capabilitytypes.Capability, error) {
	channelID, capKey, err := k.ChanOpenInit(
		ctx, msg.Channel.Ordering, msg.Channel.ConnectionHops, msg.PortId,
		portCap, msg.Channel.Counterparty, msg.Channel.Version,
	)
	if err != nil {
		return nil, "", nil, sdkerrors.Wrap(err, "channel handshake open init failed")
	}

	if err := ctx.EventManager().EmitTypedEvent(
		&types.EventChannelOpenInit{
			PortId:             msg.PortId,
			ChannelId:          channelID,
			CounterpartyPortId: msg.Channel.Counterparty.PortId,
			ConnectionId:       msg.Channel.ConnectionHops[0],
		},
	); err != nil {
		return nil, "", nil, err
	}

	return &sdk.Result{
		Events: ctx.EventManager().Events().ToABCIEvents(),
	}, channelID, capKey, nil
}

// HandleMsgChannelOpenTry defines the sdk.Handler for MsgChannelOpenTry
func HandleMsgChannelOpenTry(ctx sdk.Context, k keeper.Keeper, portCap *capabilitytypes.Capability, msg *types.MsgChannelOpenTry) (*sdk.Result, string, *capabilitytypes.Capability, error) {
	channelID, capKey, err := k.ChanOpenTry(ctx, msg.Channel.Ordering, msg.Channel.ConnectionHops, msg.PortId, msg.PreviousChannelId,
		portCap, msg.Channel.Counterparty, msg.Channel.Version, msg.CounterpartyVersion, msg.ProofInit, msg.ProofHeight,
	)
	if err != nil {
		return nil, "", nil, sdkerrors.Wrap(err, "channel handshake open try failed")
	}

	if err := ctx.EventManager().EmitTypedEvent(
		&types.EventChannelOpenTry{
			PortId:                msg.PortId,
			ChannelId:             channelID,
			CounterpartyPortId:    msg.Channel.Counterparty.PortId,
			CounterpartyChannelId: msg.Channel.Counterparty.ChannelId,
			ConnectionId:          msg.Channel.ConnectionHops[0],
		},
	); err != nil {
		return nil, "", nil, err
	}

	return &sdk.Result{
		Events: ctx.EventManager().Events().ToABCIEvents(),
	}, channelID, capKey, nil
}

// HandleMsgChannelOpenAck defines the sdk.Handler for MsgChannelOpenAck
func HandleMsgChannelOpenAck(ctx sdk.Context, k keeper.Keeper, channelCap *capabilitytypes.Capability, msg *types.MsgChannelOpenAck) (*sdk.Result, error) {
	err := k.ChanOpenAck(
		ctx, msg.PortId, msg.ChannelId, channelCap, msg.CounterpartyVersion, msg.CounterpartyChannelId, msg.ProofTry, msg.ProofHeight,
	)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "channel handshake open ack failed")
	}

	channel, _ := k.GetChannel(ctx, msg.PortId, msg.ChannelId)

	if err := ctx.EventManager().EmitTypedEvent(
		&types.EventChannelOpenAck{
			PortId:                msg.PortId,
			ChannelId:             msg.ChannelId,
			CounterpartyPortId:    channel.Counterparty.PortId,
			CounterpartyChannelId: channel.Counterparty.ChannelId,
			ConnectionId:          channel.ConnectionHops[0],
		},
	); err != nil {
		return nil, err
	}

	return &sdk.Result{
		Events: ctx.EventManager().Events().ToABCIEvents(),
	}, nil
}

// HandleMsgChannelOpenConfirm defines the sdk.Handler for MsgChannelOpenConfirm
func HandleMsgChannelOpenConfirm(ctx sdk.Context, k keeper.Keeper, channelCap *capabilitytypes.Capability, msg *types.MsgChannelOpenConfirm) (*sdk.Result, error) {
	err := k.ChanOpenConfirm(ctx, msg.PortId, msg.ChannelId, channelCap, msg.ProofAck, msg.ProofHeight)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "channel handshake open confirm failed")
	}

	channel, _ := k.GetChannel(ctx, msg.PortId, msg.ChannelId)

	if err := ctx.EventManager().EmitTypedEvent(
		&types.EventChannelOpenConfirm{
			PortId:                msg.PortId,
			ChannelId:             msg.ChannelId,
			CounterpartyPortId:    channel.Counterparty.PortId,
			CounterpartyChannelId: channel.Counterparty.ChannelId,
			ConnectionId:          channel.ConnectionHops[0],
		},
	); err != nil {
		return nil, err
	}

	return &sdk.Result{
		Events: ctx.EventManager().Events().ToABCIEvents(),
	}, nil
}

// HandleMsgChannelCloseInit defines the sdk.Handler for MsgChannelCloseInit
func HandleMsgChannelCloseInit(ctx sdk.Context, k keeper.Keeper, channelCap *capabilitytypes.Capability, msg *types.MsgChannelCloseInit) (*sdk.Result, error) {
	err := k.ChanCloseInit(ctx, msg.PortId, msg.ChannelId, channelCap)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "channel handshake close init failed")
	}

	channel, _ := k.GetChannel(ctx, msg.PortId, msg.ChannelId)

	if err := ctx.EventManager().EmitTypedEvent(
		&types.EventChannelCloseInit{
			PortId:                msg.PortId,
			ChannelId:             msg.ChannelId,
			CounterpartyPortId:    channel.Counterparty.PortId,
			CounterpartyChannelId: channel.Counterparty.ChannelId,
			ConnectionId:          channel.ConnectionHops[0],
		},
	); err != nil {
		return nil, err
	}

	return &sdk.Result{
		Events: ctx.EventManager().Events().ToABCIEvents(),
	}, nil
}

// HandleMsgChannelCloseConfirm defines the sdk.Handler for MsgChannelCloseConfirm
func HandleMsgChannelCloseConfirm(ctx sdk.Context, k keeper.Keeper, channelCap *capabilitytypes.Capability, msg *types.MsgChannelCloseConfirm) (*sdk.Result, error) {
	err := k.ChanCloseConfirm(ctx, msg.PortId, msg.ChannelId, channelCap, msg.ProofInit, msg.ProofHeight)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "channel handshake close confirm failed")
	}

	channel, _ := k.GetChannel(ctx, msg.PortId, msg.ChannelId)

	if err := ctx.EventManager().EmitTypedEvent(
		&types.EventChannelCloseConfirm{
			PortId:                msg.PortId,
			ChannelId:             msg.ChannelId,
			CounterpartyPortId:    channel.Counterparty.PortId,
			CounterpartyChannelId: channel.Counterparty.ChannelId,
			ConnectionId:          channel.ConnectionHops[0],
		},
	); err != nil {
		return nil, err
	}

	return &sdk.Result{
		Events: ctx.EventManager().Events().ToABCIEvents(),
	}, nil
}
