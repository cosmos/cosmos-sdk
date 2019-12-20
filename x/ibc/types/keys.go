package types

import "fmt"

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
const (
	KeyClientPrefix int = iota + 1
	KeyClientTypePrefix
	KeyConsensusStatePrefix
	KeyRootPrefix
	KeyCommiterPrefix
	KeyClientConnectionsPrefix
	KeyConnectionPrefix
	KeyChannelPrefix
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
