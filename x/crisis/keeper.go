package crisis

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Crisis keeper
type Keeper struct {
	routes      []InvarRoute
	distrKeeper DistrKeeper
}

// Keeper - create a new crisis keeper
func NewKeeper(routes []InvarRoute, distrKeeper DistrKeeper) Keeper {
	return Keeper{
		routes:      routes,
		distrKeeper: distrKeeper,
	}
}

// register routes for the
func (k *Keeper) RegisterRoute(ctx sdk.Context, invarRoute InvarRoute) {
	k.routes = append(k.routes, invarRoute)
}

// expected bank keeper
type DistrKeeper interface {
	DistributeFeePool(ctx sdk.Context, amount sdk.Coin, receiveAddr sdk.AccAddress) error
}
