package keeper

import (
	"bytes"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	client "github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	port "github.com/cosmos/cosmos-sdk/x/ibc/05-port"
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
	channel, found := k.GetChannel(ctx, packet.GetSourcePort(), packet.GetSourceChannel())
	if !found {
		return nil, sdkerrors.Wrap(types.ErrChannelNotFound, packet.GetSourceChannel())
	}

	if channel.State != types.OPEN {
		return nil, sdkerrors.Wrapf(
			types.ErrInvalidChannelState,
			"channel state is not OPEN (got %s)", channel.State.String(),
		)
	}

	_, found = k.GetChannelCapability(ctx, packet.GetSourcePort(), packet.GetSourceChannel())
	if !found {
		return nil, types.ErrChannelCapabilityNotFound
	}

	if !k.portKeeper.Authenticate(portCapability, packet.GetSourcePort()) {
		return nil, sdkerrors.Wrapf(port.ErrInvalidPort, "invalid source port: %s", packet.GetSourcePort())
	}

	if packet.GetDestChannel() != channel.Counterparty.ChannelID {
		return nil, sdkerrors.Wrapf(
			types.ErrInvalidPacket,
			"packet destination channel doesn't match the counterparty's channel (%s ≠ %s)", packet.GetDestChannel(), channel.Counterparty.ChannelID,
		)
	}

	connectionEnd, found := k.connectionKeeper.GetConnection(ctx, channel.ConnectionHops[0])
	if !found {
		return nil, sdkerrors.Wrap(connection.ErrConnectionNotFound, channel.ConnectionHops[0])
	}

	if packet.GetDestPort() != channel.Counterparty.PortID {
		return nil, sdkerrors.Wrapf(types.ErrInvalidPacket,
			"packet destination port doesn't match the counterparty's port (%s ≠ %s)", packet.GetDestPort(), channel.Counterparty.PortID,
		)
	}

	if nextSequenceRecv >= packet.GetSequence() {
		return nil, sdkerrors.Wrap(types.ErrInvalidPacket, "packet already received")
	}

	commitment := k.GetPacketCommitment(ctx, packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())
	if !bytes.Equal(commitment, types.CommitPacket(packet.GetData())) {
		return nil, sdkerrors.Wrap(types.ErrInvalidPacket, "packet hasn't been sent")
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
		ok = k.connectionKeeper.VerifyMembership(
			ctx, connectionEnd, proofHeight, proof,
			types.PacketAcknowledgementPath(packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence()),
			acknowledgement,
		)
	default:
		panic(sdkerrors.Wrapf(types.ErrInvalidChannelOrdering, channel.Ordering.String()))
	}

	if !ok {
		return nil, sdkerrors.Wrap(types.ErrInvalidPacket, "packet verification failed")
	}

	k.deletePacketCommitment(ctx, packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())
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

	channel, found := k.GetChannel(ctx, packet.GetSourcePort(), packet.GetSourceChannel())
	if !found {
		return sdkerrors.Wrap(types.ErrChannelNotFound, packet.GetSourceChannel())
	}

	if channel.State == types.CLOSED {
		return sdkerrors.Wrapf(
			types.ErrInvalidChannelState,
			"channel is CLOSED (got %s)", channel.State.String(),
		)
	}

	if !k.portKeeper.Authenticate(portCapability, packet.GetSourcePort()) {
		return sdkerrors.Wrap(port.ErrInvalidPort, packet.GetSourcePort())
	}

	if packet.GetDestPort() != channel.Counterparty.PortID {
		return sdkerrors.Wrapf(
			types.ErrInvalidPacket,
			"packet destination port doesn't match the counterparty's port (%s ≠ %s)", packet.GetDestPort(), channel.Counterparty.PortID,
		)
	}

	if packet.GetDestChannel() != channel.Counterparty.ChannelID {
		return sdkerrors.Wrapf(
			types.ErrInvalidPacket,
			"packet destination channel doesn't match the counterparty's channel (%s ≠ %s)", packet.GetDestChannel(), channel.Counterparty.ChannelID,
		)
	}

	connectionEnd, found := k.connectionKeeper.GetConnection(ctx, channel.ConnectionHops[0])
	if !found {
		return sdkerrors.Wrap(connection.ErrConnectionNotFound, channel.ConnectionHops[0])
	}

	if connectionEnd.State == connection.UNINITIALIZED {
		return sdkerrors.Wrap(
			connection.ErrInvalidConnectionState,
			"connection is closed (i.e NONE)",
		)
	}

	_, found = k.clientKeeper.GetConsensusState(ctx, connectionEnd.ClientID)
	if !found {
		return client.ErrConsensusStateNotFound
	}

	if uint64(ctx.BlockHeight()) >= packet.GetTimeoutHeight() {
		return types.ErrPacketTimeout
	}

	nextSequenceSend, found := k.GetNextSequenceSend(ctx, packet.GetSourcePort(), packet.GetSourceChannel())
	if !found {
		return types.ErrSequenceSendNotFound
	}

	if packet.GetSequence() != nextSequenceSend {
		return sdkerrors.Wrapf(
			types.ErrInvalidPacket,
			"packet sequence ≠ next send sequence (%d ≠ %d)", packet.GetSequence(), nextSequenceSend,
		)
	}

	nextSequenceSend++
	k.SetNextSequenceSend(ctx, packet.GetSourcePort(), packet.GetSourceChannel(), nextSequenceSend)
	k.SetPacketCommitment(ctx, packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence(), types.CommitPacket(packet.GetData()))

	return nil
}

// RecvPacket is called by a module in order to receive & process an IBC packet
// sent on the corresponding channel end on the counterparty chain.
func (k Keeper) RecvPacket(
	ctx sdk.Context,
	packet exported.PacketI,
	proof commitment.ProofI,
	proofHeight uint64,
) (exported.PacketI, error) {

	channel, found := k.GetChannel(ctx, packet.GetDestPort(), packet.GetDestChannel())
	if !found {
		return nil, sdkerrors.Wrap(types.ErrChannelNotFound, packet.GetDestChannel())
	}

	if channel.State != types.OPEN {
		return nil, sdkerrors.Wrapf(
			types.ErrInvalidChannelState,
			"channel state is not OPEN (got %s)", channel.State.String(),
		)
	}

	// RecvPacket is called by the antehandler which acts upon the packet.Route(),
	// so the capability authentication can be omitted here

	// packet must come from the channel's counterparty
	if packet.GetSourcePort() != channel.Counterparty.PortID {
		return nil, sdkerrors.Wrapf(
			types.ErrInvalidPacket,
			"packet source port doesn't match the counterparty's port (%s ≠ %s)", packet.GetSourcePort(), channel.Counterparty.PortID,
		)
	}

	if packet.GetSourceChannel() != channel.Counterparty.ChannelID {
		return nil, sdkerrors.Wrapf(
			types.ErrInvalidPacket,
			"packet source channel doesn't match the counterparty's channel (%s ≠ %s)", packet.GetSourceChannel(), channel.Counterparty.ChannelID,
		)
	}

	connectionEnd, found := k.connectionKeeper.GetConnection(ctx, channel.ConnectionHops[0])
	if !found {
		return nil, sdkerrors.Wrap(connection.ErrConnectionNotFound, channel.ConnectionHops[0])
	}

	if connectionEnd.State != connection.OPEN {
		return nil, sdkerrors.Wrapf(
			connection.ErrInvalidConnectionState,
			"connection state is not OPEN (got %s)", connectionEnd.State.String(),
		)
	}

	if channel.Ordering == types.ORDERED {
		nextSequenceRecv, found := k.GetNextSequenceRecv(ctx, packet.GetDestPort(), packet.GetDestChannel())
		if !found {
			return nil, types.ErrSequenceReceiveNotFound
		}

		if packet.GetSequence() != nextSequenceRecv {
			return nil, sdkerrors.Wrapf(
				types.ErrInvalidPacket,
				"packet sequence ≠ next receive sequence (%d ≠ %d)", packet.GetSequence(), nextSequenceRecv,
			)
		}
	}

	if uint64(ctx.BlockHeight()) >= packet.GetTimeoutHeight() {
		return nil, types.ErrPacketTimeout
	}

	if !k.connectionKeeper.VerifyMembership(
		ctx, connectionEnd, proofHeight, proof,
		types.PacketCommitmentPath(packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence()),
		types.CommitPacket(packet.GetData()),
	) {
		return nil, errors.New("couldn't verify counterparty packet commitment")
	}

	return packet, nil
}

// PacketExecuted writes the packet execution acknowledgement to the state,
// which will be verified by the counterparty chain using AcknowledgePacket.
// CONTRACT: each packet handler function should call WriteAcknowledgement at the end of the execution
func (k Keeper) PacketExecuted(
	ctx sdk.Context,
	packet exported.PacketI,
	acknowledgement exported.PacketDataI,
) error {
	channel, found := k.GetChannel(ctx, packet.GetDestPort(), packet.GetDestChannel())
	if !found {
		return sdkerrors.Wrapf(types.ErrChannelNotFound, packet.GetDestChannel())
	}

	if channel.State != types.OPEN {
		return sdkerrors.Wrapf(
			types.ErrInvalidChannelState,
			"channel state is not OPEN (got %s)", channel.State.String(),
		)
	}

	if acknowledgement != nil || channel.Ordering == types.UNORDERED {
		k.SetPacketAcknowledgement(
			ctx, packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence(),
			types.CommitAcknowledgement(acknowledgement),
		)
	}

	if channel.Ordering == types.ORDERED {
		nextSequenceRecv, found := k.GetNextSequenceRecv(ctx, packet.GetDestPort(), packet.GetDestChannel())
		if !found {
			return types.ErrSequenceReceiveNotFound
		}

		if packet.GetSequence() != nextSequenceRecv {
			return sdkerrors.Wrapf(
				types.ErrInvalidPacket,
				"packet sequence ≠ next receive sequence (%d ≠ %d)", packet.GetSequence(), nextSequenceRecv,
			)
		}

		nextSequenceRecv++

		k.SetNextSequenceRecv(ctx, packet.GetDestPort(), packet.GetDestChannel(), nextSequenceRecv)
	}

	return nil
}

// AcknowledgePacket is called by a module to process the acknowledgement of a
// packet previously sent by the calling module on a channel to a counterparty
// module on the counterparty chain. acknowledgePacket also cleans up the packet
// commitment, which is no longer necessary since the packet has been received
// and acted upon.
func (k Keeper) AcknowledgePacket(
	ctx sdk.Context,
	packet exported.PacketI,
	acknowledgement exported.PacketDataI,
	proof commitment.ProofI,
	proofHeight uint64,
) (exported.PacketI, error) {
	channel, found := k.GetChannel(ctx, packet.GetSourcePort(), packet.GetSourceChannel())
	if !found {
		return nil, sdkerrors.Wrap(types.ErrChannelNotFound, packet.GetSourceChannel())
	}

	if channel.State != types.OPEN {
		return nil, sdkerrors.Wrapf(
			types.ErrInvalidChannelState,
			"channel state is not OPEN (got %s)", channel.State.String(),
		)
	}

	// RecvPacket is called by the antehandler which acts upon the packet.Route(),
	// so the capability authentication can be omitted here

	// packet must come from the channel's counterparty
	if packet.GetSourcePort() != channel.Counterparty.PortID {
		return nil, sdkerrors.Wrapf(
			types.ErrInvalidPacket,
			"packet source port doesn't match the counterparty's port (%s ≠ %s)", packet.GetSourcePort(), channel.Counterparty.PortID,
		)
	}

	if packet.GetSourceChannel() != channel.Counterparty.ChannelID {
		return nil, sdkerrors.Wrapf(
			types.ErrInvalidPacket,
			"packet source channel doesn't match the counterparty's channel (%s ≠ %s)", packet.GetSourceChannel(), channel.Counterparty.ChannelID,
		)
	}

	connectionEnd, found := k.connectionKeeper.GetConnection(ctx, channel.ConnectionHops[0])
	if !found {
		return nil, sdkerrors.Wrap(connection.ErrConnectionNotFound, channel.ConnectionHops[0])
	}

	if connectionEnd.State != connection.OPEN {
		return nil, sdkerrors.Wrapf(
			connection.ErrInvalidConnectionState,
			"connection state is not OPEN (got %s)", connectionEnd.State.String(),
		)
	}

	commitment := k.GetPacketCommitment(ctx, packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())
	if !bytes.Equal(commitment, types.CommitPacket(packet.GetData())) {
		return nil, sdkerrors.Wrap(types.ErrInvalidPacket, "packet hasn't been sent")
	}

	if !k.connectionKeeper.VerifyMembership(
		ctx, connectionEnd, proofHeight, proof,
		types.PacketAcknowledgementPath(packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence()),
		acknowledgement.GetBytes(),
	) {
		return nil, errors.New("invalid acknowledgement on counterparty chain")
	}

	return packet, nil
}

// AcknowledgementExecuted deletes the commitment send from this chain after it receives the acknowlegement
// CONTRACT: each acknowledgement handler function should call WriteAcknowledgement at the end of the execution
func (k Keeper) AcknowledgementExecuted(ctx sdk.Context, packet exported.PacketI) {
	k.deletePacketCommitment(ctx, packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())
}
