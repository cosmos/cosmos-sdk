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

	connection, found := k.connectionKeeper.GetConnection(ctx, channel.ConnectionHops[0])
	if !found {
		return nil, errors.New("connection not found") // TODO: ics03 sdk.Error
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

	connection, found := k.connectionKeeper.GetConnection(ctx, channel.ConnectionHops[0])
	if !found {
		return nil, errors.New("connection not found") // TODO: ics03 sdk.Error
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

// function sendPacket(packet: Packet) {
// 	channel = provableStore.get(channelPath(packet.sourcePort, packet.sourceChannel))

// 	// optimistic sends are permitted once the handshake has started
// 	abortTransactionUnless(channel !== null)
// 	abortTransactionUnless(channel.state !== CLOSED)
// 	abortTransactionUnless(authenticate(privateStore.get(channelCapabilityPath(packet.sourcePort, packet.sourceChannel))))
// 	abortTransactionUnless(packet.destPort === channel.counterpartyPortIdentifier)
// 	abortTransactionUnless(packet.destChannel === channel.counterpartyChannelIdentifier)
// 	connection = provableStore.get(connectionPath(channel.connectionHops[0]))

// 	abortTransactionUnless(connection !== null)
// 	abortTransactionUnless(connection.state !== CLOSED)

// 	consensusState = provableStore.get(consensusStatePath(connection.clientIdentifier))
// 	abortTransactionUnless(consensusState.getHeight() < packet.timeoutHeight)

// 	nextSequenceSend = provableStore.get(nextSequenceSendPath(packet.sourcePort, packet.sourceChannel))
// 	abortTransactionUnless(packet.sequence === nextSequenceSend)

// 	// all assertions passed, we can alter state

// 	nextSequenceSend = nextSequenceSend + 1
// 	provableStore.set(nextSequenceSendPath(packet.sourcePort, packet.sourceChannel), nextSequenceSend)
// 	provableStore.set(packetCommitmentPath(packet.sourcePort, packet.sourceChannel, packet.sequence), hash(packet.data))
// }

// function recvPacket(
//   packet: OpaquePacket,
//   proof: CommitmentProof,
//   proofHeight: uint64,
//   acknowledgement: bytes): Packet {

//     channel = provableStore.get(channelPath(packet.destPort, packet.destChannel))
//     abortTransactionUnless(channel !== null)
//     abortTransactionUnless(channel.state === OPEN)
//     abortTransactionUnless(authenticate(privateStore.get(channelCapabilityPath(packet.destPort, packet.destChannel))))
//     abortTransactionUnless(packet.sourcePort === channel.counterpartyPortIdentifier)
//     abortTransactionUnless(packet.sourceChannel === channel.counterpartyChannelIdentifier)

//     connection = provableStore.get(connectionPath(channel.connectionHops[0]))
//     abortTransactionUnless(connection !== null)
//     abortTransactionUnless(connection.state === OPEN)

//     abortTransactionUnless(getConsensusHeight() < packet.timeoutHeight)

//     abortTransactionUnless(connection.verifyMembership(
//       proofHeight,
//       proof,
//       packetCommitmentPath(packet.sourcePort, packet.sourceChannel, packet.sequence),
//       hash(packet.data)
//     ))

//     // all assertions passed (except sequence check), we can alter state

//     if (acknowledgement.length > 0 || channel.order === UNORDERED)
//       provableStore.set(
//         packetAcknowledgementPath(packet.destPort, packet.destChannel, packet.sequence),
//         hash(acknowledgement)
//       )

//     if (channel.order === ORDERED) {
//       nextSequenceRecv = provableStore.get(nextSequenceRecvPath(packet.destPort, packet.destChannel))
//       abortTransactionUnless(packet.sequence === nextSequenceRecv)
//       nextSequenceRecv = nextSequenceRecv + 1
//       provableStore.set(nextSequenceRecvPath(packet.destPort, packet.destChannel), nextSequenceRecv)
//     }

//     // return transparent packet
//     return packet
// }

func (k Keeper) SendPacket(ctx sdk.Context, channelID string, packet exported.PacketI) error {
	// obj, err := man.Query(ctx, packet.SenderPort(), chanId)
	// if err != nil {
	// 	return err
	// }

	// if obj.OriginConnection().Client.GetConsensusState(ctx).GetHeight() >= packet.Timeout() {
	// 	return errors.New("timeout height higher than the latest known")
	// }

	// obj.Packets.SetRaw(ctx, obj.SeqSend.Increment(ctx), packet.Marshal())

	return nil
}

func (k Keeper) RecvPacket(ctx sdk.Context, proofs []ics23.Proof, height uint64, portID, channelID string, packet exported.PacketI) error {
	// obj, err := man.Query(ctx, portid, chanid)
	// if err != nil {
	// 	return err
	// }

	// /*
	// 	if !obj.Receivable(ctx) {
	// 		return errors.New("cannot receive Packets on this channel")
	// 	}
	// */

	// ctx, err = obj.Context(ctx, proofs, height)
	// if err != nil {
	// 	return err
	// }

	// err = assertTimeout(ctx, packet.Timeout())
	// if err != nil {
	// 	return err
	// }

	// if !obj.counterParty.Packets.Value(obj.SeqRecv.Increment(ctx)).IsRaw(ctx, packet.Marshal()) {
	// 	return errors.New("verification failed")
	// }

	return nil
}
