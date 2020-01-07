package keeper

import (
	"bytes"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

// TimeoutPacket is called by a module which originally attempted to send a
// packet to a counterparty module, where the timeout height has passed on the
// counterparty chain without the packet being committed, to prove that the
// packet can no longer be executed and to allow the calling module to safely
// perform appropriate state transitions.
func (k Keeper) TimeoutPacket(
	ctx sdk.Context,
	packet types.Packet,
	proof commitment.ProofI,
	proofHeight uint64,
	nextSequenceRecv uint64,
) (types.Packet, error) {
	channel, found := k.GetChannel(ctx, packet.GetSourcePort(), packet.GetSourceChannel())
	if !found {
		return types.Packet{}, sdkerrors.Wrapf(
			types.ErrChannelNotFound,
			packet.GetSourcePort(), packet.GetSourceChannel(),
		)
	}

	if channel.State != types.OPEN {
		return types.Packet{}, sdkerrors.Wrapf(
			types.ErrInvalidChannelState,
			"channel state is not OPEN (got %s)", channel.State.String(),
		)
	}

	_, found = k.GetChannelCapability(ctx, packet.GetSourcePort(), packet.GetSourceChannel())
	if !found {
		return types.Packet{}, types.ErrChannelCapabilityNotFound
	}

	if packet.GetDestChannel() != channel.Counterparty.ChannelID {
		return types.Packet{}, sdkerrors.Wrapf(
			types.ErrInvalidPacket,
			"packet destination channel doesn't match the counterparty's channel (%s ≠ %s)", packet.GetDestChannel(), channel.Counterparty.ChannelID,
		)
	}

	connectionEnd, found := k.connectionKeeper.GetConnection(ctx, channel.ConnectionHops[0])
	if !found {
		return types.Packet{}, sdkerrors.Wrapf(
			connection.ErrConnectionNotFound,
			channel.ConnectionHops[0],
		)
	}

	if packet.GetDestPort() != channel.Counterparty.PortID {
		return types.Packet{}, sdkerrors.Wrapf(
			types.ErrInvalidPacket,
			"packet destination port doesn't match the counterparty's port (%s ≠ %s)", packet.GetDestPort(), channel.Counterparty.PortID,
		)
	}

	if proofHeight < packet.GetTimeoutHeight() {
		return types.Packet{}, types.ErrPacketTimeout
	}

	if nextSequenceRecv >= packet.GetSequence() {
		return types.Packet{}, sdkerrors.Wrapf(
			types.ErrInvalidPacket,
			"packet already received",
		)
	}

	commitment := k.GetPacketCommitment(ctx, packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())
	if !bytes.Equal(commitment, types.CommitPacket(packet.PacketDataI)) {
		return types.Packet{}, sdkerrors.Wrapf(
			types.ErrInvalidPacket,
			"packet hasn't been sent",
		)
	}

	var ok bool
	switch channel.Ordering {
	case types.ORDERED:
		ok = k.connectionKeeper.VerifyMembership(
			ctx, connectionEnd, proofHeight, proof,
			types.NextSequenceRecvPath(packet.GetDestPort(), packet.GetDestChannel()),
			sdk.Uint64ToBigEndian(nextSequenceRecv),
		)
	case types.UNORDERED:
		ok = k.connectionKeeper.VerifyNonMembership(
			ctx, connectionEnd, proofHeight, proof,
			types.PacketAcknowledgementPath(packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence()),
		)
	default:
		panic(sdkerrors.Wrapf(types.ErrInvalidChannelOrdering, channel.Ordering.String()))
	}

	if !ok {
		return types.Packet{}, sdkerrors.Wrap(types.ErrInvalidPacket, "packet verification failed")
	}

	k.deletePacketCommitment(ctx, packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())

	if channel.Ordering == types.ORDERED {
		channel.State = types.CLOSED
		k.SetChannel(ctx, packet.GetSourcePort(), packet.GetSourceChannel(), channel)
	}

	return packet, nil
}

// TimeoutOnClose is called by a module in order to prove that the channel to
// which an unreceived packet was addressed has been closed, so the packet will
// never be received (even if the timeoutHeight has not yet been reached).
func (k Keeper) TimeoutOnClose(
	ctx sdk.Context,
	packet types.Packet,
	proofNonMembership,
	proofClosed commitment.ProofI,
	proofHeight uint64,
	portCapability sdk.CapabilityKey,
) (types.Packet, error) {
	channel, found := k.GetChannel(ctx, packet.GetSourcePort(), packet.GetSourceChannel())
	if !found {
		return types.Packet{}, sdkerrors.Wrapf(types.ErrChannelNotFound, packet.GetSourcePort(), packet.GetSourceChannel())
	}

	_, found = k.GetChannelCapability(ctx, packet.GetSourcePort(), packet.GetSourceChannel())
	if !found {
		return types.Packet{}, types.ErrChannelCapabilityNotFound
	}

	if !k.portKeeper.Authenticate(portCapability, packet.GetSourcePort()) {
		return types.Packet{}, errors.New("port is not valid")
	}

	if packet.GetDestChannel() != channel.Counterparty.ChannelID {
		return types.Packet{}, sdkerrors.Wrapf(
			types.ErrInvalidPacket,
			"packet destination channel doesn't match the counterparty's channel (%s ≠ %s)", packet.GetDestChannel(), channel.Counterparty.ChannelID,
		)
	}

	connectionEnd, found := k.connectionKeeper.GetConnection(ctx, channel.ConnectionHops[0])
	if !found {
		return types.Packet{}, sdkerrors.Wrap(connection.ErrConnectionNotFound, channel.ConnectionHops[0])
	}

	if packet.GetDestPort() != channel.Counterparty.PortID {
		return types.Packet{}, sdkerrors.Wrapf(
			types.ErrInvalidPacket,
			"packet destination port doesn't match the counterparty's port (%s ≠ %s)", packet.GetDestPort(), channel.Counterparty.PortID,
		)
	}

	commitment := k.GetPacketCommitment(ctx, packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())
	if !bytes.Equal(commitment, types.CommitPacket(packet.PacketDataI)) {
		return types.Packet{}, sdkerrors.Wrapf(types.ErrInvalidPacket, "packet hasn't been sent")
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

	bz, err := k.cdc.MarshalBinaryLengthPrefixed(expectedChannel)
	if err != nil {
		return types.Packet{}, errors.New("failed to marshal expected channel")
	}

	if !k.connectionKeeper.VerifyMembership(
		ctx, connectionEnd, proofHeight, proofClosed,
		types.ChannelPath(channel.Counterparty.PortID, channel.Counterparty.ChannelID),
		bz,
	) {
		return types.Packet{}, sdkerrors.Wrap(
			types.ErrInvalidCounterparty,
			"channel membership verification failed",
		)
	}

	if !k.connectionKeeper.VerifyNonMembership(
		ctx, connectionEnd, proofHeight, proofNonMembership,
		types.PacketAcknowledgementPath(packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence()),
	) {
		return types.Packet{}, errors.New("cannot verify absence of acknowledgement at packet index")
	}

	k.deletePacketCommitment(ctx, packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())

	return packet, nil
}
