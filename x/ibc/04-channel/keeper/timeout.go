package keeper

import (
	"bytes"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	ics23 "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

// TimoutPacket  is called by a module which originally attempted to send a
// packet to a counterparty module, where the timeout height has passed on the
// counterparty chain without the packet being committed, to prove that the
// packet can no longer be executed and to allow the calling module to safely
// perform appropriate state transitions.
func (k Keeper) TimoutPacket(
	ctx sdk.Context,
	packet exported.PacketI,
	proof ics23.Proof,
	proofHeight uint64,
	nextSequenceRecv uint64,
) (exported.PacketI, error) {
	channel, found := k.GetChannel(ctx, packet.SourcePort(), packet.SourceChannel())
	if !found {
		return nil, errors.New("channel not found") // TODO: sdk.Error
	}

	if channel.State != types.OPEN {
		return nil, errors.New("channel is not open") // TODO: sdk.Error
	}

	_, found = k.GetChannelCapability(ctx, packet.SourcePort(), packet.SourceChannel())
	if !found {
		return nil, errors.New("channel capability key not found") // TODO: sdk.Error
	}

	// if !AuthenticateCapabilityKey(capabilityKey) {
	//  return errors.New("invalid capability key") // TODO: sdk.Error
	// }

	if packet.DestChannel() != channel.Counterparty.ChannelID {
		return nil, errors.New("invalid packet destination channel")
	}

	connection, found := k.connectionKeeper.GetConnection(ctx, channel.ConnectionHops[0])
	if !found {
		return nil, errors.New("connection not found") // TODO: ics03 sdk.Error
	}

	if packet.DestPort() != channel.Counterparty.PortID {
		return nil, errors.New("invalid packet destination port")
	}

	if proofHeight < packet.TimeoutHeight() {
		return nil, errors.New("timeout on counterparty connection end")
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
		ok = k.connectionKeeper.VerifyNonMembership(
			ctx, connection, proofHeight, proof,
			types.PacketAcknowledgementPath(packet.SourcePort(), packet.SourceChannel(), packet.Sequence()),
		)
	default:
		panic("invalid channel ordering type")
	}

	if !ok {
		return nil, errors.New("failed packet verification") // TODO: sdk.Error
	}

	k.deletePacketCommitment(ctx, packet.SourcePort(), packet.SourceChannel(), packet.Sequence())

	if channel.Ordering == types.ORDERED {
		channel.State = types.CLOSED
		k.SetChannel(ctx, packet.SourcePort(), packet.SourceChannel(), channel)
	}

	return packet, nil
}

// TimeoutOnClose is called by a module in order to prove that the channel to
// which an unreceived packet was addressed has been closed, so the packet will
// never be received (even if the timeoutHeight has not yet been reached).
func (k Keeper) TimeoutOnClose(
	ctx sdk.Context,
	packet exported.PacketI,
	proofNonMembership,
	proofClosed ics23.Proof,
	proofHeight uint64,
) (exported.PacketI, error) {
	channel, found := k.GetChannel(ctx, packet.SourcePort(), packet.SourceChannel())
	if !found {
		return nil, errors.New("channel not found") // TODO: sdk.Error
	}

	_, found = k.GetChannelCapability(ctx, packet.SourcePort(), packet.SourceChannel())
	if !found {
		return nil, errors.New("channel capability key not found") // TODO: sdk.Error
	}

	// if !AuthenticateCapabilityKey(capabilityKey) {
	//  return errors.New("invalid capability key") // TODO: sdk.Error
	// }

	if packet.DestChannel() != channel.Counterparty.ChannelID {
		return nil, errors.New("invalid packet destination channel")
	}

	connection, found := k.connectionKeeper.GetConnection(ctx, channel.ConnectionHops[0])
	if !found {
		return nil, errors.New("connection not found") // TODO: ics03 sdk.Error
	}

	if packet.DestPort() != channel.Counterparty.PortID {
		return nil, errors.New("invalid packet destination port")
	}

	if packet.DestPort() != channel.Counterparty.PortID {
		return nil, errors.New("port id doesn't match with counterparty's")
	}

	commitment := k.GetPacketCommitment(ctx, packet.SourcePort(), packet.SourceChannel(), packet.Sequence())
	if !bytes.Equal(commitment, packet.Data()) { // TODO: hash packet data
		return nil, errors.New("packet hasn't been sent")
	}

	counterparty := types.NewCounterparty(packet.SourcePort(), packet.SourceChannel())
	expectedChannel := types.NewChannel(
		types.CLOSED, channel.Ordering, counterparty, channel.CounterpartyHops(), channel.Version,
	)

	bz, err := k.cdc.MarshalBinaryLengthPrefixed(expectedChannel)
	if err != nil {
		return nil, errors.New("failed to marshal expected channel")
	}

	if !k.connectionKeeper.VerifyMembership(
		ctx, connection, proofHeight, proofClosed,
		types.ChannelPath(channel.Counterparty.PortID, channel.Counterparty.ChannelID),
		bz,
	) {
		return nil, errors.New("counterparty channel doesn't match the expected one")
	}

	if !k.connectionKeeper.VerifyNonMembership(
		ctx, connection, proofHeight, proofNonMembership,
		types.PacketAcknowledgementPath(packet.SourcePort(), packet.SourceChannel(), packet.Sequence()),
	) {
		return nil, errors.New("cannot verify absence of acknowledgement at packet index")
	}

	k.deletePacketCommitment(ctx, packet.SourcePort(), packet.SourceChannel(), packet.Sequence())

	return packet, nil
}
