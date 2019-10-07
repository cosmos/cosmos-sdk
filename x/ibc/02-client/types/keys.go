package types

import (
	"fmt"
)

const (
	// SubModuleName defines the IBC client name
	SubModuleName = "clients"
)

// The following paths are the keys to the store as defined in https://github.com/cosmos/ics/tree/master/spec/ics-002-client-semantics#path-space

// clientStatePath takes an Identifier and returns a Path under which to store a
// particular client state
func clientStatePath(clientID string) string {
	return fmt.Sprintf("clients/%s/state", clientID)
}

// clientTypePath takes an Identifier and returns Path under which to store the
// type of a particular client.
func clientTypePath(clientID string) string {
	return fmt.Sprintf("clients/%s/type", clientID)
}

// consensusStatePath takes an Identifier and returns a Path under which to
// store the consensus state of a client.
func consensusStatePath(clientID string) string {
	return fmt.Sprintf("clients/%s/consensusState", clientID)
}

// consensusStatePath takes an Identifier and returns a Path under which to
// store the consensus state of a client.
func rootPath(clientID string, height uint64) string {
	return fmt.Sprintf("clients/%s/roots/%d", clientID, height)
}

// KeyClientState returns the store key for a particular client state
func KeyClientState(clientID string) []byte {
	return []byte(clientStatePath(clientID))
}

// KeyClientType returns the store key for type of a particular client
func KeyClientType(clientID string) []byte {
	return []byte(clientTypePath(clientID))
}

// KeyConsensusState returns the store key for the consensus state of a particular
// client
func KeyConsensusState(clientID string) []byte {
	return []byte(consensusStatePath(clientID))
}

// KeyRoot returns the store key for a commitment root of a particular
// client at a given height
func KeyRoot(clientID string, height uint64) []byte {
	return []byte(rootPath(clientID, height))
}
