package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SupplyKeeper expected supply keeper
type SupplyKeeper interface {
	InflateSupply(ctx sdk.Context, supplyType string, amount sdk.Coins)
}
