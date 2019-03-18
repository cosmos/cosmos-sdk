package crisis

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// Crisis keeper
type Keeper struct {
	routes      []InvarRoute
	paramSpace  params.Subspace
	distrKeeper DistrKeeper
}

// Keeper creates a new crisis Keeper object
func NewKeeper(routes []InvarRoute, paramSpace params.Subspace, distrKeeper DistrKeeper) Keeper {
	return Keeper{
		routes:      routes,
		paramSpace:  paramSpace,
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
