package types

import "cosmossdk.io/collections"

const (
	// ModuleName defines the module name
	ModuleName = "account_link"
)

// KVStore keys
var (
	AccountsPrefix           = collections.NewPrefix(0)
	AccountTypeAddressPrefix = collections.NewPrefix(1)
)
