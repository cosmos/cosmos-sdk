package bank

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CrisisKeeper expected crisis keeper
type CrisisKeeper interface {
	RegisterRoute(moduleName, route string, invar sdk.Invariant)
}
