package types

import (
	"fmt"
)

const (
	// SubModuleName defines the IBC channels name
	SubModuleName = "channels" // TODO: why was this "ports" beforehand

	// StoreKey is the store key string for IBC channels
	StoreKey = SubModuleName

	// RouterKey is the message route for IBC channels
	RouterKey = SubModuleName

	// QuerierRoute is the querier route for IBC channels
	QuerierRoute = SubModuleName
)

func channelPath(portID string, channelID string) string {
	return fmt.Sprintf("ports/%s/channels/%s", portID, channelID)
}

func channelCapabilityPath(portID string, channelID string) string {
	return fmt.Sprintf("%s/key", channelPath(portID, channelID))
}

func nextSequenceSendPath(portID string, channelID string) string {
	return fmt.Sprintf("%s/nextSequenceSend", channelPath(portID, channelID))
}

func nextSequenceRecvPath(portID string, channelID string) string {
	return fmt.Sprintf("%s/nextSequenceRecv", channelPath(portID, channelID))
}

func packetCommitmentPath(portID string, channelID string, sequence uint64) string {
	return fmt.Sprintf("%s/packets/%d", channelPath(portID, channelID), sequence)
}

func packetAcknowledgementPath(portID string, channelID string, sequence uint64) string {
	return fmt.Sprintf("%s/acknowledgements/%d", channelPath(portID, channelID), sequence)
}

// KeyChannel returns the store key for a particular channel
func KeyChannel(portID, channelID string) []byte {
	return []byte(channelPath(portID, channelID))
}

// KeyChannelCapabilityPath returns the store key for the capability key of a
// particular channel binded to a specific port
func KeyChannelCapabilityPath(portID, channelID string) []byte {
	return []byte(channelCapabilityPath(portID, channelID))
}

// KeyNextSequenceSend returns the store key the send sequence of a particular
// channel binded to a specific port
func KeyNextSequenceSend(portID, channelID string) []byte {
	return []byte(nextSequenceSendPath(portID, channelID))
}

// KeyNextSequenceRecv returns the store key the receive sequence of a particular
// channel binded to a specific port
func KeyNextSequenceRecv(portID, channelID string) []byte {
	return []byte(nextSequenceRecvPath(portID, channelID))
}
