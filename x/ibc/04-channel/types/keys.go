package types

import (
	"fmt"
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
func ChannelPath(portID string, channelID string) string {
	return fmt.Sprintf("ports/%s/channels/%s", portID, channelID)
}

// ChannelCapabilityPath defines the path under which capability keys associated
// with a channel are stores
func ChannelCapabilityPath(portID string, channelID string) string {
	return fmt.Sprintf("%s/key", ChannelPath(portID, channelID))
}

// NextSequenceSendPath defines the next send sequence counter store path
func NextSequenceSendPath(portID string, channelID string) string {
	return fmt.Sprintf("%s/nextSequenceSend", ChannelPath(portID, channelID))
}

// NextSequenceRecvPath defines the next receive sequence counter store path
func NextSequenceRecvPath(portID string, channelID string) string {
	return fmt.Sprintf("%s/nextSequenceRecv", ChannelPath(portID, channelID))
}

// PacketCommitmentPath defines the commitments to packet data fields store path
func PacketCommitmentPath(portID string, channelID string, sequence uint64) string {
	return fmt.Sprintf("%s/packets/%d", ChannelPath(portID, channelID), sequence)
}

// PacketAcknowledgementPath defines the packet acknowledgement store path
func PacketAcknowledgementPath(portID string, channelID string, sequence uint64) string {
	return fmt.Sprintf("%s/acknowledgements/%d", ChannelPath(portID, channelID), sequence)
}

// KeyChannel returns the store key for a particular channel
func KeyChannel(portID, channelID string) []byte {
	return []byte(ChannelPath(portID, channelID))
}

// KeyChannelCapabilityPath returns the store key for the capability key of a
// particular channel binded to a specific port
func KeyChannelCapabilityPath(portID, channelID string) []byte {
	return []byte(ChannelCapabilityPath(portID, channelID))
}

// KeyNextSequenceSend returns the store key the send sequence of a particular
// channel binded to a specific port
func KeyNextSequenceSend(portID, channelID string) []byte {
	return []byte(NextSequenceSendPath(portID, channelID))
}

// KeyNextSequenceRecv returns the store key the receive sequence of a particular
// channel binded to a specific port
func KeyNextSequenceRecv(portID, channelID string) []byte {
	return []byte(NextSequenceRecvPath(portID, channelID))
}
