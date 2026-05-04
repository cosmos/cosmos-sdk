package types

import (
	"cosmossdk.io/collections"
)

const (
	ModuleName = "auth"

	// StoreKey is string representation of the store key for auth
	StoreKey = "acc"

	// FeeCollectorName the root string for the fee collector account address
	FeeCollectorName = "fee_collector"
)

var (
	// ParamsKey is the prefix for params key
	ParamsKey = collections.NewPrefix(0)

	// AddressStoreKeyPrefix prefix for account-by-address store
	AddressStoreKeyPrefix = collections.NewPrefix(1)

	// Deprecated: GlobalAccountNumberKey identifies the prefix where the monotonically increasing
	// account number is stored.
	// This storage is deprecated as of 0.45, but left here to prevent future use of the ID
	GlobalAccountNumberKey = collections.NewPrefix(2)

	// AccountNumberStoreKeyPrefix prefix for account-by-id store
	AccountNumberStoreKeyPrefix = collections.NewPrefix("accountNumber")

	// UnorderedNoncesKey prefix for the unordered sequence storage.
	UnorderedNoncesKey = collections.NewPrefix(90)

	// LegacyGlobalAccountNumberKey is the legacy param key for global account number
	LegacyGlobalAccountNumberKey = []byte("globalAccountNumber")
)
