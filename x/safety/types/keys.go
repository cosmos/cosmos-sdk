package types

import "cosmossdk.io/collections"

const (
	// ModuleName defines the module name.
	ModuleName = "safety"

	// StoreKey defines the primary module store key.
	StoreKey = ModuleName
)

var (
	// ParamsKey is the prefix for the module params.
	ParamsKey = collections.NewPrefix(0)

	// DisabledListKey is the prefix for the set of disabled patterns.
	DisabledListKey = collections.NewPrefix(1)
)
