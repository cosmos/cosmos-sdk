package types

import "cosmossdk.io/collections"

const (
	// ModuleName defines the module name
	ModuleName = "circuit"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName
)

// KVStore keys
var (
	AccountPermissionPrefix = collections.NewPrefix(1)
	DisableListPrefix       = collections.NewPrefix(2)
)
