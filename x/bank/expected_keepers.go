package bank

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CrisisKeeper expected crisis keeper
type CrisisKeeper interface {
	RegisterRoute(moduleName, route string, invar sdk.Invariant)
}

// SupplyKeeper expected supply keeper
type SupplyKeeper interface {
	InflateSupply(ctx sdk.Context, supplyType string, amount sdk.Coins)
	DeflateSupply(ctx sdk.Context, supplyType string, amount sdk.Coins)
}
