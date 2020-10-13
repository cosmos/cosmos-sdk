package types

const (
	// ModuleName defines the IBC convo name
	ModuleName = "convo"

	// Version defines the current version the IBC convo
	// module supports
	Version = "convo-v1"

	// PortID is the default port id that transfer module binds to
	PortID = "conversation"

	// StoreKey is the store key string for IBC transfer
	StoreKey = ModuleName

	// RouterKey is the message route for IBC transfer
	RouterKey = ModuleName

	// QuerierRoute is the querier route for IBC transfer
	QuerierRoute = ModuleName
)

var (
	// PortKey defines the key to store the port ID in store
	PortKey = []byte{0x01}
)
