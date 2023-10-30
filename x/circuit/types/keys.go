package types

import "cosmossdk.io/collections"

// ModuleName defines the module name
const ModuleName = "circuit"

// KVStore keys
var (
	AccountPermissionPrefix = collections.NewPrefix(1)
	DisableListPrefix       = collections.NewPrefix(2)
)
