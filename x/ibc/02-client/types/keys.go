package types

import (
	"fmt"
)

const (
	// SubModuleName defines the IBC client name
	SubModuleName = "client"

	// StoreKey is the store key string for IBC client
	StoreKey = SubModuleName

	// RouterKey is the message route for IBC client
	RouterKey = SubModuleName

	// QuerierRoute is the querier route for IBC client
	QuerierRoute = SubModuleName
)

// The following paths are the keys to the store as defined in https://github.com/cosmos/ics/tree/master/spec/ics-002-client-semantics#path-space

// ClientStatePath takes an Identifier and returns a Path under which to store a
// particular client state
func ClientStatePath(clientID string) string {
	return fmt.Sprintf("clients/%s/state", clientID)
}

// ClientTypePath takes an Identifier and returns Path under which to store the
// type of a particular client.
func ClientTypePath(clientID string) string {
	return fmt.Sprintf("clients/%s/type", clientID)
}

// ConsensusStatePath takes an Identifier and returns a Path under which to
// store the consensus state of a client.
func ConsensusStatePath(clientID string) string {
	return fmt.Sprintf("clients/%s/consensusState", clientID)
}

// RootPath takes an Identifier and returns a Path under which to
// store the consensus state of a client.
func RootPath(clientID string, height uint64) string {
	return fmt.Sprintf("clients/%s/roots/%d", clientID, height)
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
func KeyConsensusState(clientID string) []byte {
	return []byte(ConsensusStatePath(clientID))
}

// KeyRoot returns the store key for a commitment root of a particular
// client at a given height
func KeyRoot(clientID string, height uint64) []byte {
	return []byte(RootPath(clientID, height))
}
