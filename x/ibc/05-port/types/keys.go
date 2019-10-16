package types

import (
	"fmt"
)

const (
	// SubModuleName defines the IBC port name
	SubModuleName = "port"

	// StoreKey is the store key string for IBC ports
	StoreKey = SubModuleName

	// RouterKey is the message route for IBC ports
	RouterKey = SubModuleName

	// QuerierRoute is the querier route for IBC ports
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
