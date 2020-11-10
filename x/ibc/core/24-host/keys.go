package host

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
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
	KeyClientStorePrefix = []byte("clients")
)

// KVStore key prefixes for IBC
const (
	KeyClientState             = "clientState"
	KeyConsensusStatePrefix    = "consensusStates"
	KeyConnectionPrefix        = "connections"
	KeyChannelPrefix           = "channelEnds"
	KeyChannelCapabilityPrefix = "capabilities"
	KeyNextSeqSendPrefix       = "nextSequenceSend"
	KeyNextSeqRecvPrefix       = "nextSequenceRecv"
	KeyNextSeqAckPrefix        = "nextSequenceAck"
	KeyPacketCommitmentPrefix  = "commitments"
	KeyPacketAckPrefix         = "acks"
	KeyPacketReceiptPrefix     = "receipts"
)

// FullClientPath returns the full path of a specific client path in the format:
// "clients/{clientID}/{path}" as a string.
func FullClientPath(clientID string, path string) string {
	return fmt.Sprintf("%s/%s/%s", KeyClientStorePrefix, clientID, path)
}

// FullClientKey returns the full path of specific client path in the format:
// "clients/{clientID}/{path}" as a byte array.
func FullClientKey(clientID string, path []byte) []byte {
	return []byte(FullClientPath(clientID, string(path)))
}

// ICS02
// The following paths are the keys to the store as defined in https://github.com/cosmos/ics/tree/master/spec/ics-002-client-semantics#path-space

// ClientStatePath takes a client identifier and returns a Path under which to store a
// particular client state
func ClientStatePath(clientID string) string {
	return FullClientPath(clientID, KeyClientState)
}

// ClientStateKey takes a client identifier and returns a Key under which to store a
// particular client state.
func ClientStateKey(clientID string) []byte {
	return FullClientKey(clientID, []byte(KeyClientState))
}

// ConsensusStatePath takes a client identifier and returns a Path under which to
// store the consensus state of a client.
func ConsensusStatePath(clientID string, height exported.Height) string {
	return FullClientPath(clientID, fmt.Sprintf("%s/%s", KeyConsensusStatePrefix, height))
}

// ConsensusStateKey returns the store key for the consensus state of a particular
// client.
func ConsensusStateKey(clientID string, height exported.Height) []byte {
	return []byte(ConsensusStatePath(clientID, height))
}

// ICS03
// The following paths are the keys to the store as defined in https://github.com/cosmos/ics/tree/master/spec/ics-003-connection-semantics#store-paths

// ClientConnectionsPath defines a reverse mapping from clients to a set of connections
func ClientConnectionsPath(clientID string) string {
	return FullClientPath(clientID, KeyConnectionPrefix)
}

// ClientConnectionsKey returns the store key for the connections of a given client
func ClientConnectionsKey(clientID string) []byte {
	return []byte(ClientConnectionsPath(clientID))
}

// ConnectionPath defines the path under which connection paths are stored
func ConnectionPath(connectionID string) string {
	return fmt.Sprintf("%s/%s", KeyConnectionPrefix, connectionID)
}

// ConnectionKey returns the store key for a particular connection
func ConnectionKey(connectionID string) []byte {
	return []byte(ConnectionPath(connectionID))
}

// ICS04
// The following paths are the keys to the store as defined in https://github.com/cosmos/ics/tree/master/spec/ics-004-channel-and-packet-semantics#store-paths

// ChannelPath defines the path under which channels are stored
func ChannelPath(portID, channelID string) string {
	return fmt.Sprintf("%s/%s", KeyChannelPrefix, channelPath(portID, channelID))
}

// KeyChannel returns the store key for a particular channel
func KeyChannel(portID, channelID string) []byte {
	return []byte(ChannelPath(portID, channelID))
}

// ChannelCapabilityPath defines the path under which capability keys associated
// with a channel are stored
func ChannelCapabilityPath(portID, channelID string) string {
	return fmt.Sprintf("%s/", KeyChannelCapabilityPrefix) + channelPath(portID, channelID)
}

// NextSequenceSendPath defines the next send sequence counter store path
func NextSequenceSendPath(portID, channelID string) string {
	return fmt.Sprintf("%s/%s", KeyNextSeqSendPrefix, channelPath(portID, channelID))
}

// NextSequenceSendKey returns the store key for the send sequence of a particular
// channel binded to a specific port.
func NextSequenceSendKey(portID, channelID string) []byte {
	return []byte(NextSequenceSendPath(portID, channelID))
}

// NextSequenceRecvPath defines the next receive sequence counter store path.
func NextSequenceRecvPath(portID, channelID string) string {
	return fmt.Sprintf("%s/%s", KeyNextSeqRecvPrefix, channelPath(portID, channelID))
}

// NextSequenceRecvKey returns the store key for the receive sequence of a particular
// channel binded to a specific port
func NextSequenceRecvKey(portID, channelID string) []byte {
	return []byte(NextSequenceRecvPath(portID, channelID))
}

// NextSequenceAckPath defines the next acknowledgement sequence counter store path
func NextSequenceAckPath(portID, channelID string) string {
	return fmt.Sprintf("%s/%s", KeyNextSeqAckPrefix, channelPath(portID, channelID))
}

// NextSequenceAckKey returns the store key for the acknowledgement sequence of
// a particular channel binded to a specific port.
func NextSequenceAckKey(portID, channelID string) []byte {
	return []byte(NextSequenceAckPath(portID, channelID))
}

// PacketCommitmentPath defines the commitments to packet data fields store path
func PacketCommitmentPath(portID, channelID string, sequence uint64) string {
	return fmt.Sprintf("%s/%s", KeyPacketCommitmentPrefix, channelPath(portID, channelID)+fmt.Sprintf("//%d", sequence))
}

// PacketCommitmentPrefixPath defines the prefix for commitments to packet data fields store path.
func PacketCommitmentPrefixPath(portID, channelID string) string {
	return fmt.Sprintf("%s/", KeyPacketCommitmentPrefix) + channelPath(portID, channelID)
}

// PacketAcknowledgementPath defines the packet acknowledgement store path
func PacketAcknowledgementPath(portID, channelID string, sequence uint64) string {
	return fmt.Sprintf("%s/", KeyPacketAckPrefix) + channelPath(portID, channelID) + fmt.Sprintf("/acknowledgements/%d", sequence)
}

// PacketAcknowledgementPrefixPath defines the prefix for commitments to packet data fields store path.
func PacketAcknowledgementPrefixPath(portID, channelID string) string {
	return fmt.Sprintf("%s/", KeyPacketAckPrefix) + channelPath(portID, channelID)
}

// PacketReceiptPath defines the packet receipt store path
func PacketReceiptPath(portID, channelID string, sequence uint64) string {
	return fmt.Sprintf("%s/", KeyPacketReceiptPrefix) + channelPath(portID, channelID) + fmt.Sprintf("/receipts/%d", sequence)
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

// KeyPacketReceipt returns the store key of under which a packet
// receipt is stored
func KeyPacketReceipt(portID, channelID string, sequence uint64) []byte {
	return []byte(PacketReceiptPath(portID, channelID, sequence))
}

func channelPath(portID, channelID string) string {
	return fmt.Sprintf("ports/%s/channels/%s", portID, channelID)
}

// ICS05
// The following paths are the keys to the store as defined in https://github.com/cosmos/ics/tree/master/spec/ics-005-port-allocation#store-paths

// PortPath defines the path under which ports paths are stored on the capability module
func PortPath(portID string) string {
	return fmt.Sprintf("ports/%s", portID)
}
