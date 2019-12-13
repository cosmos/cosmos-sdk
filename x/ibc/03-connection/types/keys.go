package types

import (
	"fmt"

	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

const (
	// SubModuleName defines the IBC connection name
	SubModuleName = "connection"

	// StoreKey is the store key string for IBC connections
	StoreKey = SubModuleName

	// RouterKey is the message route for IBC connections
	RouterKey = SubModuleName

	// QuerierRoute is the querier route for IBC connections
	QuerierRoute = SubModuleName
)

// The following paths are the keys to the store as defined in https://github.com/cosmos/ics/tree/master/spec/ics-003-connection-semantics#store-paths

// ClientConnectionsPath defines a reverse mapping from clients to a set of connections
func ClientConnectionsPath(clientID string) string {
	return fmt.Sprintf("clients/%s/connections", clientID)
}

// ConnectionPath defines the path under which connection paths are stored
func ConnectionPath(connectionID string) string {
	return fmt.Sprintf("connections/%s", connectionID)
}

// KeyClientConnections returns the store key for the connectios of a given client
func KeyClientConnections(clientID string) []byte {
	return append(
		ibctypes.KeyPrefixBytes(ibctypes.KeyClientConnectionsPrefix),
		[]byte(ClientConnectionsPath(clientID))...,
	)
}

// KeyConnection returns the store key for a particular connection
func KeyConnection(connectionID string) []byte {
	return append(
		ibctypes.KeyPrefixBytes(ibctypes.KeyConnectionPrefix),
		[]byte(ConnectionPath(connectionID))...,
	)
}

// GetConnectionsKeysPrefix return the connection prefixes
func GetConnectionsKeysPrefix(prefix int) []byte {
	return []byte(fmt.Sprintf("%d/connections", prefix))
}
