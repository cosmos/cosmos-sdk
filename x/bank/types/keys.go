package types

import (
	"cosmossdk.io/collections"
)

const (
	// ModuleName defines the module name
	ModuleName = "bank"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName
)

// KVStore keys
var (
	SupplyKey           = collections.NewPrefix(0)
	DenomMetadataPrefix = collections.NewPrefix(1)
	// BalancesPrefix is the prefix for the account balances store. We use a byte
	// (instead of `[]byte("balances")` to save some disk space).
	BalancesPrefix     = collections.NewPrefix(2)
	DenomAddressPrefix = collections.NewPrefix(3)
	// SendEnabledPrefix is the prefix for the SendDisabled flags for a Denom.
	SendEnabledPrefix = collections.NewPrefix(4)

	// ParamsKey is the prefix for x/bank parameters
	ParamsKey = collections.NewPrefix(5)
)
