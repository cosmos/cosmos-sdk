package types

import (
	"fmt"
	"strings"
)

const (
	// ModuleName is the name of the IBC module
	ModuleName = "ibc"

	// StoreKey is the string store representation
	StoreKey string = ModuleName

	// QuerierRoute is the querier route for the IBC module
	QuerierRoute string = ModuleName

	// RouterKey is the msg router key for the IBC module
	RouterKey string = ModuleName
)

// KVStore key prefixes for IBC
var (
	KeyClientPrefix            = []byte("clientState")
	KeyClientTypePrefix        = []byte("clientType")
	KeyConsensusStatePrefix    = []byte("consensusState")
	KeyClientConnectionsPrefix = []byte("clientConnections")
	KeyConnectionPrefix        = []byte("connection")
)

// KVStore key prefixes for IBC
const (
	KeyChannelPrefix int = iota + 1
	KeyChannelCapabilityPrefix
	KeyNextSeqSendPrefix
	KeyNextSeqRecvPrefix
	KeyPacketCommitmentPrefix
	KeyPacketAckPrefix
	KeyPortsPrefix
)

// KeyPrefixBytes return the key prefix bytes from a URL string format
func KeyPrefixBytes(prefix int) []byte {
	return []byte(fmt.Sprintf("%d/", prefix))
}

// ICS02
// The following paths are the keys to the store as defined in https://github.com/cosmos/ics/tree/master/spec/ics-002-client-semantics#path-space

// ClientStatePath takes an Identifier and returns a Path under which to store a
// particular client state
func ClientStatePath(clientID string) string {
	return fmt.Sprintf("clientState/%s", clientID)
}

// ClientTypePath takes an Identifier and returns Path under which to store the
// type of a particular client.
func ClientTypePath(clientID string) string {
	return fmt.Sprintf("clientType/%s", clientID)
}

// ConsensusStatePath takes an Identifier and returns a Path under which to
// store the consensus state of a client.
func ConsensusStatePath(clientID string, height uint64) string {
	return fmt.Sprintf("consensusState/%s/%d", clientID, height)
}

// KeyClientState returns the store key for a particular client state
func KeyClientState(clientID string) []byte {
	return []byte(ClientStatePath(clientID))
}

// KeyClientType returns the store key for type of a particular client
func KeyClientType(clientID string) []byte {
	return []byte(ClientTypePath(clientID))
}

// KeyConsensusState returns the store key for the consensus state of a particular
// client
func KeyConsensusState(clientID string, height uint64) []byte {
	return []byte(ConsensusStatePath(clientID, height))
}

// ICS03
// The following paths are the keys to the store as defined in https://github.com/cosmos/ics/tree/master/spec/ics-003-connection-semantics#store-paths

// ClientConnectionsPath defines a reverse mapping from clients to a set of connections
func ClientConnectionsPath(clientID string) string {
	return fmt.Sprintf("clientConnections/%s/", clientID)
}

// ConnectionPath defines the path under which connection paths are stored
func ConnectionPath(connectionID string) string {
	return fmt.Sprintf("connection/%s", connectionID)
}

// KeyClientConnections returns the store key for the connectios of a given client
func KeyClientConnections(clientID string) []byte {
	return []byte(ClientConnectionsPath(clientID))
}

// KeyConnection returns the store key for a particular connection
func KeyConnection(connectionID string) []byte {
	return []byte(ConnectionPath(connectionID))
}

// ICS04
// The following paths are the keys to the store as defined in https://github.com/cosmos/ics/tree/master/spec/ics-004-channel-and-packet-semantics#store-paths

// GetChannelPortsKeysPrefix returns the prefix bytes for ICS04 and ICS05 iterators
func GetChannelPortsKeysPrefix(prefix int) []byte {
	return []byte(fmt.Sprintf("%d/ports/", prefix))
}

// ChannelPath defines the path under which channels are stored
func ChannelPath(portID, channelID string) string {
	return fmt.Sprintf("%d/", KeyChannelPrefix) + channelPath(portID, channelID)
}

// ChannelCapabilityPath defines the path under which capability keys associated
// with a channel are stored
func ChannelCapabilityPath(portID, channelID string) string {
	return fmt.Sprintf("%d/", KeyChannelCapabilityPrefix) + channelPath(portID, channelID) + "/key"
}

// NextSequenceSendPath defines the next send sequence counter store path
func NextSequenceSendPath(portID, channelID string) string {
	return fmt.Sprintf("%d/", KeyNextSeqSendPrefix) + channelPath(portID, channelID) + "/nextSequenceSend"
}

// NextSequenceRecvPath defines the next receive sequence counter store path
func NextSequenceRecvPath(portID, channelID string) string {
	return fmt.Sprintf("%d/", KeyNextSeqRecvPrefix) + channelPath(portID, channelID) + "/nextSequenceRecv"
}

// PacketCommitmentPath defines the commitments to packet data fields store path
func PacketCommitmentPath(portID, channelID string, sequence uint64) string {
	return fmt.Sprintf("%d/", KeyPacketCommitmentPrefix) + channelPath(portID, channelID) + fmt.Sprintf("/packets/%d", sequence)
}

// PacketAcknowledgementPath defines the packet acknowledgement store path
func PacketAcknowledgementPath(portID, channelID string, sequence uint64) string {
	return fmt.Sprintf("%d/", KeyPacketAckPrefix) + channelPath(portID, channelID) + fmt.Sprintf("/acknowledgements/%d", sequence)
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

// KeyNextSequenceSend returns the store key for the send sequence of a particular
// channel binded to a specific port
func KeyNextSequenceSend(portID, channelID string) []byte {
	return []byte(NextSequenceSendPath(portID, channelID))
}

// KeyNextSequenceRecv returns the store key for the receive sequence of a particular
// channel binded to a specific port
func KeyNextSequenceRecv(portID, channelID string) []byte {
	return []byte(NextSequenceRecvPath(portID, channelID))
}

// KeyPacketCommitment returns the store key of under which a packet commitment
// is stored
func KeyPacketCommitment(portID, channelID string, sequence uint64) []byte {
	return []byte(PacketCommitmentPath(portID, channelID, sequence))
}

// KeyPacketAcknowledgement returns the store key of under which a packet
// acknowledgement is stored
func KeyPacketAcknowledgement(portID, channelID string, sequence uint64) []byte {
	return []byte(PacketAcknowledgementPath(portID, channelID, sequence))
}

func channelPath(portID, channelID string) string {
	return fmt.Sprintf("ports/%s/channels/%s", portID, channelID)
}

func MustParseChannelPath(path string) (string, string) {
	split := strings.Split(path, "/")
	if len(split) != 5 {
		panic("cannot parse channel path")
	}
	if split[1] != "ports" || split[3] != "channels" {
		panic("cannot parse channel path")
	}
	return split[2], split[4]
}

// ICS05
// The following paths are the keys to the store as defined in https://github.com/cosmos/ics/tree/master/spec/ics-005-port-allocation#store-paths

// PortPath defines the path under which ports paths are stored
func PortPath(portID string) string {
	return fmt.Sprintf("%d/ports/%s", KeyPortsPrefix, portID)
}

// KeyPort returns the store key for a particular port
func KeyPort(portID string) []byte {
	return []byte(PortPath(portID))
}
