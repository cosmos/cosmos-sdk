package types

import (
	"fmt"

	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

const (
	// SubModuleName defines the IBC channels name
	SubModuleName = "channels"

	// StoreKey is the store key string for IBC channels
	StoreKey = SubModuleName

	// RouterKey is the message route for IBC channels
	RouterKey = SubModuleName

	// QuerierRoute is the querier route for IBC channels
	QuerierRoute = SubModuleName
)

// ChannelPath defines the path under which channels are stored
func ChannelPath(portID, channelID string) string {
	return string(KeyChannel(portID, channelID))
}

// ChannelCapabilityPath defines the path under which capability keys associated
// with a channel are stored
func ChannelCapabilityPath(portID, channelID string) string {
	return string(KeyChannelCapabilityPath(portID, channelID))
}

// NextSequenceSendPath defines the next send sequence counter store path
func NextSequenceSendPath(portID, channelID string) string {
	return string(KeyNextSequenceSend(portID, channelID))
}

// NextSequenceRecvPath defines the next receive sequence counter store path
func NextSequenceRecvPath(portID, channelID string) string {
	return string(KeyNextSequenceRecv(portID, channelID))
}

// PacketCommitmentPath defines the commitments to packet data fields store path
func PacketCommitmentPath(portID, channelID string, sequence uint64) string {
	return string(KeyPacketCommitment(portID, channelID, sequence))
}

// PacketAcknowledgementPath defines the packet acknowledgement store path
func PacketAcknowledgementPath(portID, channelID string, sequence uint64) string {
	return string(KeyPacketAcknowledgement(portID, channelID, sequence))
}

// KeyChannel returns the store key for a particular channel
func KeyChannel(portID, channelID string) []byte {
	return append(
		ibctypes.KeyPrefixBytes(ibctypes.KeyChannelPrefix),
		[]byte(channelPath(portID, channelID))...,
	)
}

// KeyChannelCapabilityPath returns the store key for the capability key of a
// particular channel binded to a specific port
func KeyChannelCapabilityPath(portID, channelID string) []byte {
	return append(
		ibctypes.KeyPrefixBytes(ibctypes.KeyChannelCapabilityPrefix),
		[]byte(channelCapabilityPath(portID, channelID))...,
	)
}

// KeyNextSequenceSend returns the store key for the send sequence of a particular
// channel binded to a specific port
func KeyNextSequenceSend(portID, channelID string) []byte {
	return append(
		ibctypes.KeyPrefixBytes(ibctypes.KeyNextSeqSendPrefix),
		[]byte(nextSequenceSendPath(portID, channelID))...,
	)
}

// KeyNextSequenceRecv returns the store key for the receive sequence of a particular
// channel binded to a specific port
func KeyNextSequenceRecv(portID, channelID string) []byte {
	return append(
		ibctypes.KeyPrefixBytes(ibctypes.KeyNextSeqRecvPrefix),
		[]byte(nextSequenceRecvPath(portID, channelID))...,
	)
}

// KeyPacketCommitment returns the store key of under which a packet commitment
// is stored
func KeyPacketCommitment(portID, channelID string, sequence uint64) []byte {
	return append(
		ibctypes.KeyPrefixBytes(ibctypes.KeyPacketCommitmentPrefix),
		[]byte(packetCommitmentPath(portID, channelID, sequence))...,
	)
}

// KeyPacketAcknowledgement returns the store key of under which a packet
// acknowledgement is stored
func KeyPacketAcknowledgement(portID, channelID string, sequence uint64) []byte {
	return append(
		ibctypes.KeyPrefixBytes(ibctypes.KeyPacketAckPrefix),
		[]byte(packetAcknowledgementPath(portID, channelID, sequence))...,
	)
}

// GetChannelKeysPrefix returns the prefix bytes for ICS04 iterators
func GetChannelKeysPrefix(prefix int) []byte {
	return []byte(fmt.Sprintf("%d/ports/", prefix))
}

func channelPath(portID, channelID string) string {
	return fmt.Sprintf("ports/%s/channels/%s", portID, channelID)
}

func channelCapabilityPath(portID, channelID string) string {
	return channelPath(portID, channelID) + "/key"
}

func nextSequenceSendPath(portID, channelID string) string {
	return channelPath(portID, channelID) + "/nextSequenceSend"
}

func nextSequenceRecvPath(portID, channelID string) string {
	return channelPath(portID, channelID) + "/nextSequenceRecv"
}

func packetCommitmentPath(portID, channelID string, sequence uint64) string {
	return channelPath(portID, channelID) + fmt.Sprintf("/packets/%d", sequence)
}

func packetAcknowledgementPath(portID, channelID string, sequence uint64) string {
	return channelPath(portID, channelID) + fmt.Sprintf("/acknowledgements/%d", sequence)
}
