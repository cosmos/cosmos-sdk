package keeper

import (
	"bytes"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	ics23 "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
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
	proof ics23.Proof,
	proofHeight,
	nextSequenceRecv uint64,
	acknowledgement []byte,
) (exported.PacketI, error) {
	channel, found := k.GetChannel(ctx, packet.SourcePort(), packet.SourceChannel())
	if !found {
		return nil, types.ErrChannelNotFound(k.codespace, packet.SourceChannel())
	}

	if channel.State != types.OPEN {
		return nil, errors.New("channel is not open") // TODO: sdk.Error
	}

	_, found = k.GetChannelCapability(ctx, packet.SourcePort(), packet.SourceChannel())
	if !found {
		return nil, types.ErrChannelCapabilityNotFound(k.codespace)
	}

	// if !AuthenticateCapabilityKey(capabilityKey) {
	//  return errors.New("invalid capability key") // TODO: sdk.Error
	// }

	if packet.DestChannel() != channel.Counterparty.ChannelID {
		return nil, errors.New("invalid packet destination channel")
	}

	connection, found := k.connectionKeeper.GetConnection(ctx, channel.ConnectionHops[0])
	if !found {
		return nil, connectiontypes.ErrConnectionNotFound(k.codespace, channel.ConnectionHops[0])
	}

	if packet.DestPort() != channel.Counterparty.PortID {
		return nil, errors.New("invalid packet destination port")
	}

	if nextSequenceRecv >= packet.Sequence() {
		return nil, errors.New("packet already received")
	}

	commitment := k.GetPacketCommitment(ctx, packet.SourcePort(), packet.SourceChannel(), packet.Sequence())
	if !bytes.Equal(commitment, packet.Data()) { // TODO: hash packet data
		return nil, errors.New("packet hasn't been sent")
	}

	var ok bool
	switch channel.Ordering {
	case types.ORDERED:
		ok = k.connectionKeeper.VerifyMembership(
			ctx, connection, proofHeight, proof,
			types.NextSequenceRecvPath(packet.DestPort(), packet.DestChannel()),
			sdk.Uint64ToBigEndian(nextSequenceRecv),
		)
	case types.UNORDERED:
		ok = k.connectionKeeper.VerifyMembership(
			ctx, connection, proofHeight, proof,
			types.PacketAcknowledgementPath(packet.DestPort(), packet.DestChannel(), packet.Sequence()),
			acknowledgement,
		)
	default:
		panic("invalid channel ordering type")
	}

	if !ok {
		return nil, errors.New("failed packet verification") // TODO: sdk.Error
	}

	k.deletePacketCommitment(ctx, packet.SourcePort(), packet.SourceChannel(), packet.Sequence())
	return packet, nil
}

// SendPacket  is called by a module in order to send an IBC packet on a channel
// end owned by the calling module to the corresponding module on the counterparty
// chain.
func (k Keeper) SendPacket(ctx sdk.Context, packet exported.PacketI) error {
	channel, found := k.GetChannel(ctx, packet.SourcePort(), packet.SourceChannel())
	if !found {
		return errors.New("channel not found") // TODO: sdk.Error
	}

	if channel.State == types.CLOSED {
		return errors.New("channel is closed") // TODO: sdk.Error
	}

	_, found = k.GetChannelCapability(ctx, packet.SourcePort(), packet.SourceChannel())
	if !found {
		return errors.New("channel capability key not found") // TODO: sdk.Error
	}

	// if !AuthenticateCapabilityKey(capabilityKey) {
	//  return errors.New("invalid capability key") // TODO: sdk.Error
	// }

	if packet.DestPort() != channel.Counterparty.PortID {
		return errors.New("invalid packet destination port")
	}

	if packet.DestChannel() != channel.Counterparty.ChannelID {
		return errors.New("invalid packet destination channel")
	}

	connection, found := k.connectionKeeper.GetConnection(ctx, channel.ConnectionHops[0])
	if !found {
		return errors.New("connection not found") // TODO: ics03 sdk.Error
	}

	if connection.State == connectiontypes.NONE {
		return errors.New("connection is closed") // TODO: sdk.Error
	}

	consensusState, found := k.clientKeeper.GetConsensusState(ctx, connection.ClientID)
	if !found {
		return errors.New("consensus state not found") // TODO: sdk.Error
	}

	if consensusState.GetHeight() >= packet.TimeoutHeight() {
		return types.ErrPacketTimeout(k.codespace)
	}

	nextSequenceSend, found := k.GetNextSequenceSend(ctx, packet.SourcePort(), packet.SourceChannel())
	if !found {
		return types.ErrSequenceNotFound(k.codespace, "send")
	}

	if packet.Sequence() != nextSequenceSend {
		return types.ErrInvalidPacketSequence(k.codespace)
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
	proof ics23.Proof,
	proofHeight uint64,
	acknowledgement []byte,
) (exported.PacketI, error) {

	channel, found := k.GetChannel(ctx, packet.SourcePort(), packet.SourceChannel())
	if !found {
		return nil, types.ErrChannelNotFound(k.codespace, packet.SourceChannel())
	}

	if channel.State != types.OPEN {
		return nil, errors.New("channel not open") // TODO: sdk.Error
	}

	_, found = k.GetChannelCapability(ctx, packet.SourcePort(), packet.SourceChannel())
	if !found {
		return nil, types.ErrChannelCapabilityNotFound(k.codespace)
	}

	// if !AuthenticateCapabilityKey(capabilityKey) {
	//  return errors.New("invalid capability key") // TODO: sdk.Error
	// }

	// packet must come from the channel's counterparty
	if packet.SourcePort() != channel.Counterparty.PortID {
		return nil, errors.New("invalid packet source port")
	}

	if packet.SourceChannel() != channel.Counterparty.ChannelID {
		return nil, errors.New("invalid packet source channel")
	}

	connection, found := k.connectionKeeper.GetConnection(ctx, channel.ConnectionHops[0])
	if !found {
		return nil, connectiontypes.ErrConnectionNotFound(k.codespace, channel.ConnectionHops[0])
	}

	if connection.State != connectiontypes.OPEN {
		return nil, errors.New("connection is not open") // TODO: ics03 sdk.Error
	}

	if uint64(ctx.BlockHeight()) >= packet.TimeoutHeight() {
		return nil, types.ErrPacketTimeout(k.codespace)
	}

	if !k.connectionKeeper.VerifyMembership(
		ctx, connection, proofHeight, proof,
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
			return nil, types.ErrInvalidPacketSequence(k.codespace)
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
	proof ics23.Proof,
	proofHeight uint64,
) (exported.PacketI, error) {
	channel, found := k.GetChannel(ctx, packet.SourcePort(), packet.SourceChannel())
	if !found {
		return nil, types.ErrChannelNotFound(k.codespace, packet.SourceChannel())
	}

	if channel.State != types.OPEN {
		return nil, errors.New("channel not open") // TODO: sdk.Error
	}

	_, found = k.GetChannelCapability(ctx, packet.SourcePort(), packet.SourceChannel())
	if !found {
		return nil, types.ErrChannelCapabilityNotFound(k.codespace)
	}

	// if !AuthenticateCapabilityKey(capabilityKey) {
	//  return errors.New("invalid capability key") // TODO: sdk.Error
	// }

	// packet must come from the channel's counterparty
	if packet.SourcePort() != channel.Counterparty.PortID {
		return nil, errors.New("invalid packet source port")
	}

	if packet.SourceChannel() != channel.Counterparty.ChannelID {
		return nil, errors.New("invalid packet source channel")
	}

	connection, found := k.connectionKeeper.GetConnection(ctx, channel.ConnectionHops[0])
	if !found {
		return nil, connectiontypes.ErrConnectionNotFound(k.codespace, channel.ConnectionHops[0])
	}

	if connection.State != connectiontypes.OPEN {
		return nil, errors.New("connection is not open") // TODO: ics03 sdk.Error
	}

	commitment := k.GetPacketCommitment(ctx, packet.SourcePort(), packet.SourceChannel(), packet.Sequence())
	if !bytes.Equal(commitment, packet.Data()) { // TODO: hash packet data
		return nil, errors.New("packet hasn't been sent")
	}

	if !k.connectionKeeper.VerifyMembership(
		ctx, connection, proofHeight, proof,
		types.PacketAcknowledgementPath(packet.DestPort(), packet.DestChannel(), packet.Sequence()),
		acknowledgement, // TODO: hash ACK
	) {
		return nil, errors.New("invalid acknowledgement on counterparty chain") // TODO: sdk.Error
	}

	k.deletePacketCommitment(ctx, packet.SourcePort(), packet.SourceChannel(), packet.Sequence())
	return packet, nil
}
