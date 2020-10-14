package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/ibc/applications/convo/types"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"
)

// SendConvo handles sending a message from a sender to a receiver
// over the given source channel
// It will create and send the IBC convo packet and store the pending message in state
func (k Keeper) SendConvo(
	ctx sdk.Context,
	sourcePort,
	sourceChannel string,
	sender sdk.AccAddress,
	receiver,
	message string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
) error {
	sourceChannelEnd, found := k.channelKeeper.GetChannel(ctx, sourcePort, sourceChannel)
	if !found {
		return sdkerrors.Wrapf(channeltypes.ErrChannelNotFound, "port ID (%s) channel ID (%s)", sourcePort, sourceChannel)
	}

	destinationPort := sourceChannelEnd.GetCounterparty().GetPortID()
	destinationChannel := sourceChannelEnd.GetCounterparty().GetChannelID()

	// get the next sequence
	sequence, found := k.channelKeeper.GetNextSequenceSend(ctx, sourcePort, sourceChannel)
	if !found {
		return sdkerrors.Wrapf(
			channeltypes.ErrSequenceSendNotFound,
			"source port: %s, source channel: %s", sourcePort, sourceChannel,
		)
	}

	channelCap, ok := k.scopedKeeper.GetCapability(ctx, host.ChannelCapabilityPath(sourcePort, sourceChannel))
	if !ok {
		return sdkerrors.Wrap(channeltypes.ErrChannelCapabilityNotFound, "module does not own channel capability")
	}

	if prev := k.GetPendingMessage(ctx, sender.String(), sourceChannel, receiver); prev != "" {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "cannot send a new message to receiver %s over channel %s while pending message %s has not been successfully received",
			receiver, sourceChannel, prev)
	}

	// store the message as a pending message
	k.SetPendingMessage(ctx, sender.String(), sourceChannel, receiver, message)

	// create and send ConvoPacket to IBC module
	packetData := types.NewConvoPacketData(
		sender.String(), receiver, message,
	)

	packet := channeltypes.NewPacket(
		packetData.GetBytes(),
		sequence,
		sourcePort,
		sourceChannel,
		destinationPort,
		destinationChannel,
		timeoutHeight,
		timeoutTimestamp,
	)

	return k.channelKeeper.SendPacket(ctx, channelCap, packet)
}

// OnRecvPacket processes a cross chain convo message.
// It will retrieve the message from the packet and store it in the intended inbox of the recipient
func (k Keeper) OnRecvPacket(ctx sdk.Context, packet channeltypes.Packet, data types.ConvoPacketData) error {
	// validate packet data upon receiving
	if err := data.ValidateBasic(); err != nil {
		return err
	}

	// set packet message in the appropriate inbox
	k.SetInboxMessage(ctx, data.Sender, packet.DestinationChannel, data.Receiver, data.Message)
	return nil
}

// OnAcknowledgementPacket responds to a successful message delivery to the intended recipient
// by clearing the pending message and replacing the outbox with the latest confirmed message
// If the acknowledgement failed, then the pending message is still cleared so that a new message may be sent
func (k Keeper) OnAcknowledgementPacket(ctx sdk.Context, packet channeltypes.Packet, data types.ConvoPacketData, ack channeltypes.Acknowledgement) error {
	// Regardless of acknowledgement success or failure, we must clear pending message
	// to allow for next message to be sent
	k.DeletePendingMessage(ctx, data.Sender, packet.SourceChannel, data.Receiver)
	switch ack.Response.(type) {
	case *channeltypes.Acknowledgement_Error:
		return nil
	default:
		// the acknowledgement succeeded on the receiving chain
		// so set the last confirmed outgoing message to the packet's message
		k.SetOutboxMessage(ctx, data.Sender, packet.SourceChannel, data.Receiver, data.Message)
		return nil
	}
}

// OnTimeoutPacket simply clears the pending message so that the sender can either re-send the original message or send a new one
func (k Keeper) OnTimeoutPacket(ctx sdk.Context, packet channeltypes.Packet, data types.ConvoPacketData) error {
	k.DeletePendingMessage(ctx, data.Sender, packet.SourceChannel, data.Receiver)
	return nil
}
