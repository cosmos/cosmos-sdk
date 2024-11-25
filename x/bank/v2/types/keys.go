package types

import (
	"cosmossdk.io/collections"
)

const (
	// ModuleName is the name of the module
	ModuleName = "bankv2"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// GovModuleName duplicates the gov module's name to avoid a dependency with x/gov.
	GovModuleName = "gov"

	// MintModuleName duplicates the mint module's name to avoid a cyclic dependency with x/mint.
	// It should be synced with the mint module's name if it is ever changed.
	// See: https://github.com/cosmos/cosmos-sdk/blob/0e34478eb7420b69869ed50f129fc274a97a9b06/x/mint/types/keys.go#L13
	MintModuleName = "mint"
)

var (
	// ParamsKey is the prefix for x/bank/v2 parameters
	ParamsKey = collections.NewPrefix(2)

	// BalancesPrefix is the prefix for the account balances store. We use a byte
	// (instead of `[]byte("balances")` to save some disk space).
	BalancesPrefix = collections.NewPrefix(3)

	DenomAddressPrefix = collections.NewPrefix(4)

	SupplyKey = collections.NewPrefix(5)
)
