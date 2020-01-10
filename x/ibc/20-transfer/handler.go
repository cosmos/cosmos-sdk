package transfer

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/ibc/20-transfer/types"

	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
)

// NewHandler returns sdk.Handler for IBC token transfer module messages
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
				return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized ics20 packet data type: %T", msg)
			}
		case channeltypes.MsgTimeout:
			switch data := msg.Data.(type) {
			case PacketDataTransfer:
				return handleTimeoutDataTransfer(ctx, k, msg, data)
			default:
				return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized ics20 packet data type: %T", data)
			}
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized ics20 message type: %T", msg)
		}
	}
}

func handleMsgTransfer(ctx sdk.Context, k Keeper, msg MsgTransfer) (*sdk.Result, error) {
	err := k.SendTransfer(ctx, msg.SourcePort, msg.SourceChannel, msg.Amount, msg.Sender, msg.Receiver, msg.Source)
	if err != nil {
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

func handlePacketDataTransfer(ctx sdk.Context, k Keeper, msg channeltypes.MsgPacket, data types.PacketDataTransfer) (*sdk.Result, error) {
	packet := msg.Packet
	err := k.ReceiveTransfer(ctx, packet.SourcePort, packet.SourceChannel, packet.DestinationPort, packet.DestinationChannel, data)
	if err != nil {
		panic(err)
		// TODO: Source chain sent invalid packet, shutdown channel
	}

	acknowledgement := types.AckDataTransfer{}
	err = k.PacketExecuted(ctx, packet, acknowledgement)
	if err != nil {
		panic(err)
		// TODO: This should not happen
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

func handleTimeoutDataTransfer(ctx sdk.Context, k Keeper, msg channeltypes.MsgTimeout, data types.PacketDataTransfer) (*sdk.Result, error) {
	packet := msg.Packet
	err := k.TimeoutTransfer(ctx, packet.SourcePort, packet.SourceChannel, packet.DestinationPort, packet.DestinationChannel, data)
	if err != nil {
		// This chain sent invalid packet
		panic(err)
	}
	// packet timeout should not fail
	return &sdk.Result{
		Events: ctx.EventManager().Events(),
	}, nil
}
