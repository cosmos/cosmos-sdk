package types

import "cosmossdk.io/collections"

const (
	// ModuleName defines the module name
	ModuleName = "circuit"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// GovModuleName duplicates the gov module's name to avoid a cyclic dependency with x/gov.
	// It should be synced with the gov module's name if it is ever changed.
	// See: https://github.com/cosmos/cosmos-sdk/blob/b62a28aac041829da5ded4aeacfcd7a42873d1c8/x/gov/types/keys.go#L9
	GovModuleName = "gov"
)

// KVStore keys
var (
	AccountPermissionPrefix = collections.NewPrefix(1)
	DisableListPrefix       = collections.NewPrefix(2)
)
