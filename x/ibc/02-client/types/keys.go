package types

import (
	"fmt"

	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

const (
	// SubModuleName defines the IBC client name
	SubModuleName string = "client"

	// StoreKey is the store key string for IBC client
	StoreKey string = SubModuleName

	// RouterKey is the message route for IBC client
	RouterKey string = SubModuleName

	// QuerierRoute is the querier route for IBC client
	QuerierRoute string = SubModuleName
)

// The following paths are the keys to the store as defined in https://github.com/cosmos/ics/tree/master/spec/ics-002-client-semantics#path-space

// ClientStatePath takes an Identifier and returns a Path under which to store a
// particular client state
func ClientStatePath(clientID string) string {
	return string(KeyClientState(clientID))
}

// ClientTypePath takes an Identifier and returns Path under which to store the
// type of a particular client.
func ClientTypePath(clientID string) string {
	return string(KeyClientType(clientID))
}

// ConsensusStatePath takes an Identifier and returns a Path under which to
// store the consensus state of a client.
func ConsensusStatePath(clientID string, height uint64) string {
	return string(KeyConsensusState(clientID, height))
}

// RootPath takes an Identifier and returns a Path under which to
// store the root for a particular height of a client.
func RootPath(clientID string, height uint64) string {
	return string(KeyRoot(clientID, height))
}

// KeyClientState returns the store key for a particular client state
func KeyClientState(clientID string) []byte {
	return append(
		ibctypes.KeyPrefixBytes(ibctypes.KeyClientPrefix),
		[]byte(clientStatePath(clientID))...,
	)
}

// KeyClientType returns the store key for type of a particular client
func KeyClientType(clientID string) []byte {
	return append(
		ibctypes.KeyPrefixBytes(ibctypes.KeyClientTypePrefix),
		[]byte(clientTypePath(clientID))...,
	)
}

// KeyConsensusState returns the store key for the consensus state of a particular
// client
func KeyConsensusState(clientID string, height uint64) []byte {
	return append(
		ibctypes.KeyPrefixBytes(ibctypes.KeyConsensusStatePrefix),
		[]byte(consensusStatePath(clientID, height))...,
	)
}

// KeyRoot returns the store key for a commitment root of a particular
// client at a given height
func KeyRoot(clientID string, height uint64) []byte {
	return append(
		ibctypes.KeyPrefixBytes(ibctypes.KeyRootPrefix),
		[]byte(rootPath(clientID, height))...,
	)
}

// GetClientKeysPrefix return the ICS02 prefix bytes
func GetClientKeysPrefix(prefix int) []byte {
	return []byte(fmt.Sprintf("%d/clients", prefix))
}

func clientStatePath(clientID string) string {
	return fmt.Sprintf("clients/%s/state", clientID)
}

func clientTypePath(clientID string) string {
	return fmt.Sprintf("clients/%s/type", clientID)
}

func consensusStatePath(clientID string, height uint64) string {
	return fmt.Sprintf("consensusState/%s/%d", clientID, height)
}

func rootPath(clientID string, height uint64) string {
	return fmt.Sprintf("clients/%s/roots/%d", clientID, height)
}
