package transfer

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
)

// NewHandler returns sdk.Handler for IBC token transfer module messages
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		switch msg := msg.(type) {
		case MsgTransfer:
			return handleMsgTransfer(ctx, k, msg)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized ICS-20 transfer message type: %T", msg)
		}
	}
}

// See createOutgoingPacket in spec:https://github.com/cosmos/ics/tree/master/spec/ics-020-fungible-token-transfer#packet-relay
func handleMsgTransfer(ctx sdk.Context, k Keeper, msg MsgTransfer) (*sdk.Result, error) {
	if err := k.SendTransfer(
		ctx, msg.SourcePort, msg.SourceChannel, msg.DestHeight, msg.Amount, msg.Sender, msg.Receiver,
	); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(AttributeKeyReceiver, msg.Receiver.String()),
		),
	)

	return &sdk.Result{
		Events: ctx.EventManager().Events().ToABCIEvents(),
	}, nil
}

// See onRecvPacket in spec: https://github.com/cosmos/ics/tree/master/spec/ics-020-fungible-token-transfer#packet-relay
func handlePacketDataTransfer(
	ctx sdk.Context, k Keeper, packet channeltypes.Packet, data FungibleTokenPacketData,
) (*sdk.Result, error) {
	if err := k.ReceiveTransfer(ctx, packet, data); err != nil {
		// NOTE (cwgoes): How do we want to handle this case? Maybe we should be more lenient,
		// it's safe to leave the channel open I think.

		// TODO: handle packet receipt that due to an error (specify)
		// the receiving chain couldn't process the transfer

		// source chain sent invalid packet, shutdown our channel end
		if err := k.ChanCloseInit(ctx, packet.DestinationPort, packet.DestinationChannel); err != nil {
			return nil, err
		}
		return nil, err
	}

	acknowledgement := AckDataTransfer{}
	if err := k.PacketExecuted(ctx, packet, acknowledgement.GetBytes()); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, AttributeValueCategory),
		),
	)

	return &sdk.Result{
		Events: ctx.EventManager().Events().ToABCIEvents(),
	}, nil
}

// See onTimeoutPacket in spec: https://github.com/cosmos/ics/tree/master/spec/ics-020-fungible-token-transfer#packet-relay
func handleTimeoutDataTransfer(
	ctx sdk.Context, k Keeper, packet channeltypes.Packet, data FungibleTokenPacketData,
) (*sdk.Result, error) {
	if err := k.TimeoutTransfer(ctx, packet, data); err != nil {
		// This shouldn't happen, since we've already validated that we've sent the packet.
		panic(err)
	}

	if err := k.TimeoutExecuted(ctx, packet); err != nil {
		// This shouldn't happen, since we've already validated that we've sent the packet.
		// TODO: Figure out what happens if the capability authorisation changes.
		panic(err)
	}

	return &sdk.Result{
		Events: ctx.EventManager().Events().ToABCIEvents(),
	}, nil
}
