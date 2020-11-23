package keeper

import (
	"bytes"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/core/03-connection/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
)

// TimeoutPacket is called by a module which originally attempted to send a
// packet to a counterparty module, where the timeout height has passed on the
// counterparty chain without the packet being committed, to prove that the
// packet can no longer be executed and to allow the calling module to safely
// perform appropriate state transitions. Its intended usage is within the
// ante handler.
func (k Keeper) TimeoutPacket(
	ctx sdk.Context,
	packet exported.PacketI,
	proof []byte,
	proofHeight exported.Height,
	nextSequenceRecv uint64,
) error {
	channel, found := k.GetChannel(ctx, packet.GetSourcePort(), packet.GetSourceChannel())
	if !found {
		return sdkerrors.Wrapf(
			types.ErrChannelNotFound,
			"port ID (%s) channel ID (%s)", packet.GetSourcePort(), packet.GetSourceChannel(),
		)
	}

	if channel.State != types.OPEN {
		return sdkerrors.Wrapf(
			types.ErrInvalidChannelState,
			"channel state is not OPEN (got %s)", channel.State.String(),
		)
	}

	// NOTE: TimeoutPacket is called by the AnteHandler which acts upon the packet.Route(),
	// so the capability authentication can be omitted here

	if packet.GetDestPort() != channel.Counterparty.PortId {
		return sdkerrors.Wrapf(
			types.ErrInvalidPacket,
			"packet destination port doesn't match the counterparty's port (%s ≠ %s)", packet.GetDestPort(), channel.Counterparty.PortId,
		)
	}

	if packet.GetDestChannel() != channel.Counterparty.ChannelId {
		return sdkerrors.Wrapf(
			types.ErrInvalidPacket,
			"packet destination channel doesn't match the counterparty's channel (%s ≠ %s)", packet.GetDestChannel(), channel.Counterparty.ChannelId,
		)
	}

	connectionEnd, found := k.connectionKeeper.GetConnection(ctx, channel.ConnectionHops[0])
	if !found {
		return sdkerrors.Wrap(
			connectiontypes.ErrConnectionNotFound,
			channel.ConnectionHops[0],
		)
	}

	// check that timeout height or timeout timestamp has passed on the other end
	proofTimestamp, err := k.connectionKeeper.GetTimestampAtHeight(ctx, connectionEnd, proofHeight)
	if err != nil {
		return err
	}

	timeoutHeight := packet.GetTimeoutHeight()
	if (timeoutHeight.IsZero() || proofHeight.LT(timeoutHeight)) &&
		(packet.GetTimeoutTimestamp() == 0 || proofTimestamp < packet.GetTimeoutTimestamp()) {
		return sdkerrors.Wrap(types.ErrPacketTimeout, "packet timeout has not been reached for height or timestamp")
	}

	commitment := k.GetPacketCommitment(ctx, packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())

	packetCommitment := types.CommitPacket(k.cdc, packet)

	// verify we sent the packet and haven't cleared it out yet
	if !bytes.Equal(commitment, packetCommitment) {
		return sdkerrors.Wrapf(types.ErrInvalidPacket, "packet commitment bytes are not equal: got (%v), expected (%v)", commitment, packetCommitment)
	}

	switch channel.Ordering {
	case types.ORDERED:
		// check that packet has not been received
		if nextSequenceRecv > packet.GetSequence() {
			return sdkerrors.Wrapf(
				types.ErrInvalidPacket,
				"packet already received, next sequence receive > packet sequence (%d > %d)", nextSequenceRecv, packet.GetSequence(),
			)
		}

		// check that the recv sequence is as claimed
		err = k.connectionKeeper.VerifyNextSequenceRecv(
			ctx, connectionEnd, proofHeight, proof,
			packet.GetDestPort(), packet.GetDestChannel(), nextSequenceRecv,
		)
	case types.UNORDERED:
		err = k.connectionKeeper.VerifyPacketReceiptAbsence(
			ctx, connectionEnd, proofHeight, proof,
			packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence(),
		)
	default:
		panic(sdkerrors.Wrapf(types.ErrInvalidChannelOrdering, channel.Ordering.String()))
	}

	if err != nil {
		return err
	}

	// NOTE: the remaining code is located in the TimeoutExecuted function
	return nil
}

// TimeoutExecuted deletes the commitment send from this chain after it verifies timeout.
// If the timed-out packet came from an ORDERED channel then this channel will be closed.
//
// CONTRACT: this function must be called in the IBC handler
func (k Keeper) TimeoutExecuted(
	ctx sdk.Context,
	chanCap *capabilitytypes.Capability,
	packet exported.PacketI,
) error {
	channel, found := k.GetChannel(ctx, packet.GetSourcePort(), packet.GetSourceChannel())
	if !found {
		return sdkerrors.Wrapf(types.ErrChannelNotFound, "port ID (%s) channel ID (%s)", packet.GetSourcePort(), packet.GetSourceChannel())
	}

	capName := host.ChannelCapabilityPath(packet.GetSourcePort(), packet.GetSourceChannel())
	if !k.scopedKeeper.AuthenticateCapability(ctx, chanCap, capName) {
		return sdkerrors.Wrapf(
			types.ErrChannelCapabilityNotFound,
			"caller does not own capability for channel with capability name %s", capName,
		)
	}

	k.deletePacketCommitment(ctx, packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())

	if channel.Ordering == types.ORDERED {
		channel.State = types.CLOSED
		k.SetChannel(ctx, packet.GetSourcePort(), packet.GetSourceChannel(), channel)
	}

	k.Logger(ctx).Info("packet timed-out", "packet", fmt.Sprintf("%v", packet))

	// emit an event marking that we have processed the timeout
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeTimeoutPacket,
			sdk.NewAttribute(types.AttributeKeyTimeoutHeight, packet.GetTimeoutHeight().String()),
			sdk.NewAttribute(types.AttributeKeyTimeoutTimestamp, fmt.Sprintf("%d", packet.GetTimeoutTimestamp())),
			sdk.NewAttribute(types.AttributeKeySequence, fmt.Sprintf("%d", packet.GetSequence())),
			sdk.NewAttribute(types.AttributeKeySrcPort, packet.GetSourcePort()),
			sdk.NewAttribute(types.AttributeKeySrcChannel, packet.GetSourceChannel()),
			sdk.NewAttribute(types.AttributeKeyDstPort, packet.GetDestPort()),
			sdk.NewAttribute(types.AttributeKeyDstChannel, packet.GetDestChannel()),
			sdk.NewAttribute(types.AttributeKeyChannelOrdering, channel.Ordering.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	})

	return nil
}

// TimeoutOnClose is called by a module in order to prove that the channel to
// which an unreceived packet was addressed has been closed, so the packet will
// never be received (even if the timeoutHeight has not yet been reached).
func (k Keeper) TimeoutOnClose(
	ctx sdk.Context,
	chanCap *capabilitytypes.Capability,
	packet exported.PacketI,
	proof,
	proofClosed []byte,
	proofHeight exported.Height,
	nextSequenceRecv uint64,
) error {
	channel, found := k.GetChannel(ctx, packet.GetSourcePort(), packet.GetSourceChannel())
	if !found {
		return sdkerrors.Wrapf(types.ErrChannelNotFound, "port ID (%s) channel ID (%s)", packet.GetSourcePort(), packet.GetSourceChannel())
	}

	capName := host.ChannelCapabilityPath(packet.GetSourcePort(), packet.GetSourceChannel())
	if !k.scopedKeeper.AuthenticateCapability(ctx, chanCap, capName) {
		return sdkerrors.Wrapf(
			types.ErrInvalidChannelCapability,
			"channel capability failed authentication with capability name %s", capName,
		)
	}

	if packet.GetDestPort() != channel.Counterparty.PortId {
		return sdkerrors.Wrapf(
			types.ErrInvalidPacket,
			"packet destination port doesn't match the counterparty's port (%s ≠ %s)", packet.GetDestPort(), channel.Counterparty.PortId,
		)
	}

	if packet.GetDestChannel() != channel.Counterparty.ChannelId {
		return sdkerrors.Wrapf(
			types.ErrInvalidPacket,
			"packet destination channel doesn't match the counterparty's channel (%s ≠ %s)", packet.GetDestChannel(), channel.Counterparty.ChannelId,
		)
	}

	connectionEnd, found := k.connectionKeeper.GetConnection(ctx, channel.ConnectionHops[0])
	if !found {
		return sdkerrors.Wrap(connectiontypes.ErrConnectionNotFound, channel.ConnectionHops[0])
	}

	commitment := k.GetPacketCommitment(ctx, packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())

	packetCommitment := types.CommitPacket(k.cdc, packet)

	// verify we sent the packet and haven't cleared it out yet
	if !bytes.Equal(commitment, packetCommitment) {
		return sdkerrors.Wrapf(types.ErrInvalidPacket, "packet commitment bytes are not equal: got (%v), expected (%v)", commitment, packetCommitment)
	}

	counterpartyHops, found := k.CounterpartyHops(ctx, channel)
	if !found {
		// Should not reach here, connectionEnd was able to be retrieved above
		panic("cannot find connection")
	}

	counterparty := types.NewCounterparty(packet.GetSourcePort(), packet.GetSourceChannel())
	expectedChannel := types.NewChannel(
		types.CLOSED, channel.Ordering, counterparty, counterpartyHops, channel.Version,
	)

	// check that the opposing channel end has closed
	if err := k.connectionKeeper.VerifyChannelState(
		ctx, connectionEnd, proofHeight, proofClosed,
		channel.Counterparty.PortId, channel.Counterparty.ChannelId,
		expectedChannel,
	); err != nil {
		return err
	}

	var err error
	switch channel.Ordering {
	case types.ORDERED:
		// check that packet has not been received
		if nextSequenceRecv > packet.GetSequence() {
			return sdkerrors.Wrapf(types.ErrInvalidPacket, "packet already received, next sequence receive > packet sequence (%d > %d", nextSequenceRecv, packet.GetSequence())
		}

		// check that the recv sequence is as claimed
		err = k.connectionKeeper.VerifyNextSequenceRecv(
			ctx, connectionEnd, proofHeight, proof,
			packet.GetDestPort(), packet.GetDestChannel(), nextSequenceRecv,
		)
	case types.UNORDERED:
		err = k.connectionKeeper.VerifyPacketReceiptAbsence(
			ctx, connectionEnd, proofHeight, proof,
			packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence(),
		)
	default:
		panic(sdkerrors.Wrapf(types.ErrInvalidChannelOrdering, channel.Ordering.String()))
	}

	if err != nil {
		return err
	}

	// NOTE: the remaining code is located in the TimeoutExecuted function
	return nil
}
