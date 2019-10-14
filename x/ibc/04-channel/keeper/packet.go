package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	ics23 "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

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
