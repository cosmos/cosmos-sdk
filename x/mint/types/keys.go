package types

import "cosmossdk.io/collections"

var (
	// MinterKey is the key to use for the keeper store.
	MinterKey = collections.NewPrefix(0)
)

const (
	// ModuleName is the module name
	ModuleName = "mint"

	// StoreKey is the default store key for mint
	StoreKey = ModuleName
)
