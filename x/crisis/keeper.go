package crisis

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// Keeper - crisis keeper
type Keeper struct {
	routes      []InvarRoute
	paramSpace  params.Subspace
	distrKeeper DistrKeeper
}

// NewKeeper creates a new crisis Keeper object
func NewKeeper(paramSpace params.Subspace, distrKeeper DistrKeeper) Keeper {
	return Keeper{
		routes:      []InvarRoute{},
		paramSpace:  paramSpace,
		distrKeeper: distrKeeper,
	}
}

// register routes for the
func (k *Keeper) RegisterRoute(route string, invar sdk.Invariant) {
	invarRoute := NewInvarRoute(route, invar)
	k.routes = append(k.routes, invarRoute)
}

// expected bank keeper
type DistrKeeper interface {
	DistributeFeePool(ctx sdk.Context, amount sdk.Coin, receiveAddr sdk.AccAddress) error
}
