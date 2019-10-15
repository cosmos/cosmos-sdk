package types

import (
	"fmt"
)

const (
	// SubModuleName defines the IBC ports name
	SubModuleName = "ports"

	// StoreKey is the store key string for IBC connections
	StoreKey = SubModuleName

	// RouterKey is the message route for IBC connections
	RouterKey = SubModuleName

	// QuerierRoute is the querier route for IBC connections
	QuerierRoute = SubModuleName
)

// The following paths are the keys to the store as defined in https://github.com/cosmos/ics/tree/master/spec/ics-005-port-allocation#store-paths

// PortPath defines the path under which ports paths are stored
func PortPath(portID string) string {
	return fmt.Sprintf("ports/%s", portID)
}

// KeyPort returns the store key for a particular port
func KeyPort(portID string) []byte {
	return []byte(PortPath(portID))
}
