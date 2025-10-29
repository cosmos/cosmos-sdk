// Package types defines the core data structures and constants for the upgrade module.
// It provides key management utilities for storing and retrieving upgrade plans,
// completed upgrades, and version mappings in the blockchain state.
package types

import "fmt"

const (
	// ModuleName is the name of this module
	ModuleName = "upgrade"

	// RouterKey is used to route governance proposals
	RouterKey = ModuleName

	// StoreKey is the prefix under which we store this module's data
	StoreKey = ModuleName
)

const (
	// PlanByte specifies the Byte under which a pending upgrade plan is stored in the store
	PlanByte = 0x0

	// DoneByte is a prefix to look up completed upgrade plan by name
	DoneByte = 0x1

	// VersionMapByte is a prefix to look up module names (key) and versions (value)
	VersionMapByte = 0x2

	// ProtocolVersionByte is a prefix to look up Protocol Version
	ProtocolVersionByte = 0x3

	// KeyUpgradedIBCState is the key under which upgraded ibc state is stored in the upgrade store
	KeyUpgradedIBCState = "upgradedIBCState"

	// KeyUpgradedClient is the sub-key under which upgraded client state will be stored
	KeyUpgradedClient = "upgradedClient"

	// KeyUpgradedConsState is the sub-key under which upgraded consensus state will be stored
	KeyUpgradedConsState = "upgradedConsState"
)

// PlanKey returns the storage key for the current upgrade plan.
// The key is constructed from PlanByte constant to ensure immutability.
// This key is used to store the pending upgrade plan in the module's store.
func PlanKey() []byte {
	return []byte{PlanByte}
}

// UpgradedClientKey returns the storage key for the upgraded IBC client state at the given height.
// This key is used by connecting IBC chains to verify the upgraded client state
// before upgrading their own clients. The key format is: "upgradedIBCState/{height}/upgradedClient".
func UpgradedClientKey(height int64) []byte {
	return []byte(fmt.Sprintf("%s/%d/%s", KeyUpgradedIBCState, height, KeyUpgradedClient))
}

// UpgradedConsStateKey returns the storage key for the upgraded IBC consensus state at the given height.
// This key is used by connecting IBC chains to verify the upgraded consensus state
// before upgrading their own clients. The key format is: "upgradedIBCState/{height}/upgradedConsState".
func UpgradedConsStateKey(height int64) []byte {
	return []byte(fmt.Sprintf("%s/%d/%s", KeyUpgradedIBCState, height, KeyUpgradedConsState))
}
