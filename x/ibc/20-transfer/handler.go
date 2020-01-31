package transfer

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/ibc/20-transfer/types"

	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
)

// NewHandler returns sdk.Handler for IBC token transfer module messages
// See NewHandler function in ADR15: https://github.com/cosmos/ics/tree/master/spec/ics-020-fungible-token-transfer#packet-relay
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		switch msg := msg.(type) {
		case MsgTransfer:
			return handleMsgTransfer(ctx, k, msg)
		case channeltypes.MsgPacket:
			switch data := msg.Data.(type) {
			case PacketDataTransfer: // i.e fulfills the Data interface
				return handlePacketDataTransfer(ctx, k, msg, data)
			default:
				return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized ICS-20 transfer packet data type: %T", msg)
			}
		case channeltypes.MsgTimeout:
			switch data := msg.Data.(type) {
			case PacketDataTransfer:
				return handleTimeoutDataTransfer(ctx, k, msg, data)
			default:
				return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized ICS-20 transfer packet data type: %T", data)
			}
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized ICS-20 transfer message type: %T", msg)
		}
	}
}

// See createOutgoingPacket in spec:https://github.com/cosmos/ics/tree/master/spec/ics-020-fungible-token-transfer#packet-relay
func handleMsgTransfer(ctx sdk.Context, k Keeper, msg MsgTransfer) (*sdk.Result, error) {
	if err := k.SendTransfer(
		ctx, msg.SourcePort, msg.SourceChannel, msg.Amount, msg.Sender, msg.Receiver, msg.Source,
	); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(types.AttributeKeyReceiver, msg.Receiver.String()),
		),
	)

	return &sdk.Result{
		Events: ctx.EventManager().Events(),
	}, nil
}

// See onRecvPacket in spec: https://github.com/cosmos/ics/tree/master/spec/ics-020-fungible-token-transfer#packet-relay
func handlePacketDataTransfer(
	ctx sdk.Context, k Keeper, msg channeltypes.MsgPacket, data types.PacketDataTransfer,
) (*sdk.Result, error) {
	packet := msg.Packet
	if err := k.ReceiveTransfer(
		ctx, packet.SourcePort, packet.SourceChannel, packet.DestinationPort, packet.DestinationChannel, data,
	); err != nil {
		// TODO: handle packet receipt that due to an error (specify)
		// the receiving chain couldn't process the transfer

		// source chain sent invalid packet, shutdown our channel end
		if err := k.ChanCloseInit(ctx, packet.DestinationPort, packet.DestinationChannel); err != nil {
			return nil, err
		}
		return nil, err
	}

	acknowledgement := types.AckDataTransfer{}
	if err := k.PacketExecuted(ctx, packet, acknowledgement); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Signer.String()),
		),
	)

	// packet receiving should not fail
	return &sdk.Result{
		Events: ctx.EventManager().Events(),
	}, nil
}

// See onTimeoutPacket in spec: https://github.com/cosmos/ics/tree/master/spec/ics-020-fungible-token-transfer#packet-relay
func handleTimeoutDataTransfer(ctx sdk.Context, k Keeper, msg channeltypes.MsgTimeout, data types.PacketDataTransfer) (*sdk.Result, error) {
	packet := msg.Packet
	if err := k.TimeoutTransfer(
		ctx, packet.SourcePort, packet.SourceChannel, packet.DestinationPort, packet.DestinationChannel, data,
	); err != nil {
		return nil, err
	}

	// packet timeout should not fail
	return &sdk.Result{
		Events: ctx.EventManager().Events(),
	}, nil
}
