package types

import "cosmossdk.io/collections"

const (
	// ModuleName is the name of the module
	ModuleName = "bankv2"

	// GovModuleName duplicates the gov module's name to avoid a dependency with x/gov.
	GovModuleName = "gov"
)

// ParamsKey is the prefix for x/bank/v2 parameters
var ParamsKey = collections.NewPrefix(2)
