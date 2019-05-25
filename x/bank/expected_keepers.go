package bank

import (
	sdk "github.com/YunSuk-Yeo/cosmos-sdk/types"
)

// expected crisis keeper
type CrisisKeeper interface {
	RegisterRoute(moduleName, route string, invar sdk.Invariant)
}
