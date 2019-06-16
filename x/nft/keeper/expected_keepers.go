package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CrisisKeeper defines the expected crisis keeper
type CrisisKeeper interface {
	RegisterRoute(moduleName, route string, invar sdk.Invariant)
}
