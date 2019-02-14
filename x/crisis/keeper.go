package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// invariant route
type InvarRoute struct {
	Invariant
	route string
}

// Register routes
type Keeper struct {
	routes []InvarRoute
}

// Set the last total validator power.
func (k Keeper) RegisterRoute(ctx sdk.Context, power sdk.Int) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinaryLengthPrefixed(power)
	store.Set(LastTotalPowerKey, b)
}
