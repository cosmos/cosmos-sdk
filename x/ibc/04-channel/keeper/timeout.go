package keeper

import (
	"bytes"
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
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
	packet exported.PacketI,
	proof commitment.ProofI,
	proofHeight uint64,
	nextSequenceRecv uint64,
) (exported.PacketI, error) {
	channel, found := k.GetChannel(ctx, packet.SourcePort(), packet.SourceChannel())
	if !found {
		return nil, types.ErrChannelNotFound(k.codespace, packet.SourceChannel())
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

	// if !AuthenticateCapabilityKey(capabilityKey) {
	//  return errors.New("invalid capability key") // TODO: sdk.Error
	// }

	if packet.DestChannel() != channel.Counterparty.ChannelID {
		return nil, errors.New("invalid packet destination channel")
	}

	connectionEnd, found := k.connectionKeeper.GetConnection(ctx, channel.ConnectionHops[0])
	if !found {
		return nil, connection.ErrConnectionNotFound(k.codespace, channel.ConnectionHops[0])
	}

	if packet.DestPort() != channel.Counterparty.PortID {
		return nil, errors.New("invalid packet destination port")
	}

	if proofHeight < packet.TimeoutHeight() {
		return nil, types.ErrPacketTimeout(k.codespace)
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
			ctx, connectionEnd, proofHeight, proof,
			types.NextSequenceRecvPath(packet.DestPort(), packet.DestChannel()),
			sdk.Uint64ToBigEndian(nextSequenceRecv),
		)
	case types.UNORDERED:
		ok = k.connectionKeeper.VerifyNonMembership(
			ctx, connectionEnd, proofHeight, proof,
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
	proofClosed commitment.ProofI,
	proofHeight uint64,
) (exported.PacketI, error) {
	channel, found := k.GetChannel(ctx, packet.SourcePort(), packet.SourceChannel())
	if !found {
		return nil, types.ErrChannelNotFound(k.codespace, packet.SourceChannel())
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

	connectionEnd, found := k.connectionKeeper.GetConnection(ctx, channel.ConnectionHops[0])
	if !found {
		return nil, connection.ErrConnectionNotFound(k.codespace, channel.ConnectionHops[0])
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
		ctx, connectionEnd, proofHeight, proofClosed,
		types.ChannelPath(channel.Counterparty.PortID, channel.Counterparty.ChannelID),
		bz,
	) {
		return nil, types.ErrInvalidCounterpartyChannel(k.codespace)
	}

	if !k.connectionKeeper.VerifyNonMembership(
		ctx, connectionEnd, proofHeight, proofNonMembership,
		types.PacketAcknowledgementPath(packet.SourcePort(), packet.SourceChannel(), packet.Sequence()),
	) {
		return nil, errors.New("cannot verify absence of acknowledgement at packet index")
	}

	k.deletePacketCommitment(ctx, packet.SourcePort(), packet.SourceChannel(), packet.Sequence())

	return packet, nil
}
