package keeper

import (
	"bytes"
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	client "github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

// CleanupPacket is called by a module to remove a received packet commitment
// from storage. The receiving end must have already processed the packet
// (whether regularly or past timeout).
//
// In the ORDERED channel case, CleanupPacket cleans-up a packet on an ordered
// channel by proving that the packet has been received on the other end.
//
// In the UNORDERED channel case, CleanupPacket cleans-up a packet on an
// unordered channel by proving that the associated acknowledgement has been
//written.
func (k Keeper) CleanupPacket(
	ctx sdk.Context,
	packet exported.PacketI,
	proof commitment.ProofI,
	proofHeight,
	nextSequenceRecv uint64,
	acknowledgement []byte,
	portCapability sdk.CapabilityKey,
) (exported.PacketI, error) {
	channel, found := k.GetChannel(ctx, packet.SourcePort(), packet.SourceChannel())
	if !found {
		return nil, types.ErrChannelNotFound(k.codespace, packet.SourcePort(), packet.SourceChannel())
	}

	if channel.State != types.OPEN {
		return nil, types.ErrInvalidChannelState(
			k.codespace,
			fmt.Sprintf("channel state is not OPEN (got %s)", channel.State.String()),
		)
	}

	_, found = k.GetChannelCapability(ctx, packet.SourcePort(), packet.SourceChannel())
	if !found {
		return nil, types.ErrChannelCapabilityNotFound(k.codespace)
	}

	if !k.portKeeper.Authenticate(portCapability, packet.SourcePort()) {
		return nil, errors.New("port is not valid")
	}

	if packet.DestChannel() != channel.Counterparty.ChannelID {
		return nil, types.ErrInvalidPacket(
			k.codespace,
			fmt.Sprintf("packet destination channel doesn't match the counterparty's channel (%s ≠ %s)", packet.DestChannel(), channel.Counterparty.ChannelID),
		)
	}

	connectionEnd, found := k.connectionKeeper.GetConnection(ctx, channel.ConnectionHops[0])
	if !found {
		return nil, connection.ErrConnectionNotFound(k.codespace, channel.ConnectionHops[0])
	}

	if packet.DestPort() != channel.Counterparty.PortID {
		return nil, types.ErrInvalidPacket(
			k.codespace,
			fmt.Sprintf("packet destination port doesn't match the counterparty's port (%s ≠ %s)", packet.DestPort(), channel.Counterparty.PortID),
		)
	}

	if nextSequenceRecv >= packet.Sequence() {
		return nil, types.ErrInvalidPacket(k.codespace, "packet already received")
	}

	commitment := k.GetPacketCommitment(ctx, packet.SourcePort(), packet.SourceChannel(), packet.Sequence())
	if !bytes.Equal(commitment, packet.Data()) { // TODO: hash packet data
		return nil, types.ErrInvalidPacket(k.codespace, "packet hasn't been sent")
	}

	var ok bool
	switch channel.Ordering {
	case types.ORDERED:
		ok = k.connectionKeeper.VerifyMembership(
			ctx, connectionEnd, proofHeight, proof,
			types.NextSequenceRecvPath(packet.DestPort(), packet.DestChannel()),
			sdk.Uint64ToBigEndian(nextSequenceRecv),
		)
	case types.UNORDERED:
		ok = k.connectionKeeper.VerifyMembership(
			ctx, connectionEnd, proofHeight, proof,
			types.PacketAcknowledgementPath(packet.DestPort(), packet.DestChannel(), packet.Sequence()),
			acknowledgement,
		)
	default:
		panic(fmt.Sprintf("invalid channel ordering type %v", channel.Ordering))
	}

	if !ok {
		return nil, types.ErrInvalidPacket(k.codespace, "packet verification failed")
	}

	k.deletePacketCommitment(ctx, packet.SourcePort(), packet.SourceChannel(), packet.Sequence())
	return packet, nil
}

// SendPacket  is called by a module in order to send an IBC packet on a channel
// end owned by the calling module to the corresponding module on the counterparty
// chain.
func (k Keeper) SendPacket(
	ctx sdk.Context,
	packet exported.PacketI,
	portCapability sdk.CapabilityKey,
) error {
	if err := packet.ValidateBasic(); err != nil {
		return err
	}

	channel, found := k.GetChannel(ctx, packet.SourcePort(), packet.SourceChannel())
	if !found {
		return types.ErrChannelNotFound(k.codespace, packet.SourcePort(), packet.SourceChannel())
	}

	if channel.State == types.CLOSED {
		return types.ErrInvalidChannelState(
			k.codespace,
			fmt.Sprintf("channel is CLOSED (got %s)", channel.State.String()),
		)
	}

	if !k.portKeeper.Authenticate(portCapability, packet.SourcePort()) {
		return errors.New("port is not valid")
	}

	if packet.DestPort() != channel.Counterparty.PortID {
		return types.ErrInvalidPacket(
			k.codespace,
			fmt.Sprintf("packet destination port doesn't match the counterparty's port (%s ≠ %s)", packet.DestPort(), channel.Counterparty.PortID),
		)
	}

	if packet.DestChannel() != channel.Counterparty.ChannelID {
		return types.ErrInvalidPacket(
			k.codespace,
			fmt.Sprintf("packet destination channel doesn't match the counterparty's channel (%s ≠ %s)", packet.DestChannel(), channel.Counterparty.ChannelID),
		)
	}

	connectionEnd, found := k.connectionKeeper.GetConnection(ctx, channel.ConnectionHops[0])
	if !found {
		return connection.ErrConnectionNotFound(k.codespace, channel.ConnectionHops[0])
	}

	if connectionEnd.State == connection.NONE {
		return connection.ErrInvalidConnectionState(
			k.codespace,
			fmt.Sprintf("connection is closed (i.e NONE)"),
		)
	}

	consensusState, found := k.clientKeeper.GetConsensusState(ctx, connectionEnd.ClientID)
	if !found {
		return client.ErrConsensusStateNotFound(k.codespace)
	}

	if consensusState.GetHeight() >= packet.TimeoutHeight() {
		return types.ErrPacketTimeout(k.codespace)
	}

	nextSequenceSend, found := k.GetNextSequenceSend(ctx, packet.SourcePort(), packet.SourceChannel())
	if !found {
		return types.ErrSequenceNotFound(k.codespace, "send")
	}

	if packet.Sequence() != nextSequenceSend {
		return types.ErrInvalidPacket(
			k.codespace,
			fmt.Sprintf("packet sequence ≠ next send sequence (%d ≠ %d)", packet.Sequence(), nextSequenceSend),
		)
	}

	nextSequenceSend++
	k.SetNextSequenceSend(ctx, packet.SourcePort(), packet.SourceChannel(), nextSequenceSend)
	k.SetPacketCommitment(ctx, packet.SourcePort(), packet.SourceChannel(), packet.Sequence(), packet.Data()) // TODO: hash packet data

	return nil
}

// RecvPacket is called by a module in order to receive & process an IBC packet
// sent on the corresponding channel end on the counterparty chain.
func (k Keeper) RecvPacket(
	ctx sdk.Context,
	packet exported.PacketI,
	proof commitment.ProofI,
	proofHeight uint64,
	acknowledgement []byte,
	portCapability sdk.CapabilityKey,
) (exported.PacketI, error) {

	channel, found := k.GetChannel(ctx, packet.DestPort(), packet.DestChannel())
	if !found {
		return nil, types.ErrChannelNotFound(k.codespace, packet.DestPort(), packet.DestChannel())
	}

	if channel.State != types.OPEN {
		return nil, types.ErrInvalidChannelState(
			k.codespace,
			fmt.Sprintf("channel state is not OPEN (got %s)", channel.State.String()),
		)
	}

	if !k.portKeeper.Authenticate(portCapability, packet.DestPort()) {
		return nil, errors.New("port is not valid")
	}

	// packet must come from the channel's counterparty
	if packet.SourcePort() != channel.Counterparty.PortID {
		return nil, types.ErrInvalidPacket(
			k.codespace,
			fmt.Sprintf("packet source port doesn't match the counterparty's port (%s ≠ %s)", packet.SourcePort(), channel.Counterparty.PortID),
		)
	}

	if packet.SourceChannel() != channel.Counterparty.ChannelID {
		return nil, types.ErrInvalidPacket(
			k.codespace,
			fmt.Sprintf("packet source channel doesn't match the counterparty's channel (%s ≠ %s)", packet.SourceChannel(), channel.Counterparty.ChannelID),
		)
	}

	connectionEnd, found := k.connectionKeeper.GetConnection(ctx, channel.ConnectionHops[0])
	if !found {
		return nil, connection.ErrConnectionNotFound(k.codespace, channel.ConnectionHops[0])
	}

	if connectionEnd.State != connection.OPEN {
		return nil, connection.ErrInvalidConnectionState(
			k.codespace,
			fmt.Sprintf("connection state is not OPEN (got %s)", connectionEnd.State.String()),
		)
	}

	if uint64(ctx.BlockHeight()) >= packet.TimeoutHeight() {
		return nil, types.ErrPacketTimeout(k.codespace)
	}

	if !k.connectionKeeper.VerifyMembership(
		ctx, connectionEnd, proofHeight, proof,
		types.PacketCommitmentPath(packet.SourcePort(), packet.SourceChannel(), packet.Sequence()),
		packet.Data(), // TODO: hash data
	) {
		return nil, errors.New("couldn't verify counterparty packet commitment")
	}

	if len(acknowledgement) > 0 || channel.Ordering == types.UNORDERED {
		k.SetPacketAcknowledgement(
			ctx, packet.DestPort(), packet.DestChannel(), packet.Sequence(),
			acknowledgement, // TODO: hash ACK
		)
	}

	if channel.Ordering == types.ORDERED {
		nextSequenceRecv, found := k.GetNextSequenceRecv(ctx, packet.DestPort(), packet.DestChannel())
		if !found {
			return nil, types.ErrSequenceNotFound(k.codespace, "receive")
		}

		if packet.Sequence() != nextSequenceRecv {
			return nil, types.ErrInvalidPacket(
				k.codespace,
				fmt.Sprintf("packet sequence ≠ next receive sequence (%d ≠ %d)", packet.Sequence(), nextSequenceRecv),
			)
		}

		nextSequenceRecv++
		k.SetNextSequenceRecv(ctx, packet.DestPort(), packet.DestChannel(), nextSequenceRecv)
	}

	return packet, nil
}

// AcknowledgePacket is called by a module to process the acknowledgement of a
// packet previously sent by the calling module on a channel to a counterparty
// module on the counterparty chain. acknowledgePacket also cleans up the packet
// commitment, which is no longer necessary since the packet has been received
// and acted upon.
func (k Keeper) AcknowledgePacket(
	ctx sdk.Context,
	packet exported.PacketI,
	acknowledgement []byte,
	proof commitment.ProofI,
	proofHeight uint64,
	portCapability sdk.CapabilityKey,
) (exported.PacketI, error) {
	channel, found := k.GetChannel(ctx, packet.SourcePort(), packet.SourceChannel())
	if !found {
		return nil, types.ErrChannelNotFound(k.codespace, packet.SourcePort(), packet.SourceChannel())
	}

	if channel.State != types.OPEN {
		return nil, types.ErrInvalidChannelState(
			k.codespace,
			fmt.Sprintf("channel state is not OPEN (got %s)", channel.State.String()),
		)
	}

	if !k.portKeeper.Authenticate(portCapability, packet.SourcePort()) {
		return nil, errors.New("invalid capability key")
	}

	// packet must come from the channel's counterparty
	if packet.SourcePort() != channel.Counterparty.PortID {
		return nil, types.ErrInvalidPacket(
			k.codespace,
			fmt.Sprintf("packet source port doesn't match the counterparty's port (%s ≠ %s)", packet.SourcePort(), channel.Counterparty.PortID),
		)
	}

	if packet.SourceChannel() != channel.Counterparty.ChannelID {
		return nil, types.ErrInvalidPacket(
			k.codespace,
			fmt.Sprintf("packet source channel doesn't match the counterparty's channel (%s ≠ %s)", packet.SourceChannel(), channel.Counterparty.ChannelID),
		)
	}

	connectionEnd, found := k.connectionKeeper.GetConnection(ctx, channel.ConnectionHops[0])
	if !found {
		return nil, connection.ErrConnectionNotFound(k.codespace, channel.ConnectionHops[0])
	}

	if connectionEnd.State != connection.OPEN {
		return nil, connection.ErrInvalidConnectionState(
			k.codespace,
			fmt.Sprintf("connection state is not OPEN (got %s)", connectionEnd.State.String()),
		)
	}

	commitment := k.GetPacketCommitment(ctx, packet.SourcePort(), packet.SourceChannel(), packet.Sequence())
	if !bytes.Equal(commitment, packet.Data()) { // TODO: hash packet data
		return nil, types.ErrInvalidPacket(k.codespace, "packet hasn't been sent")
	}

	if !k.connectionKeeper.VerifyMembership(
		ctx, connectionEnd, proofHeight, proof,
		types.PacketAcknowledgementPath(packet.DestPort(), packet.DestChannel(), packet.Sequence()),
		acknowledgement, // TODO: hash ACK
	) {
		return nil, errors.New("invalid acknowledgement on counterparty chain")
	}

	k.deletePacketCommitment(ctx, packet.SourcePort(), packet.SourceChannel(), packet.Sequence())
	return packet, nil
}
