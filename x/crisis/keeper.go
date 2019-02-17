package crisis

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// invariant route
type InvarRoute struct {
	route string
	Invariant
}

// Register routes
type Keeper struct {
	routes []InvarRoute
}

// register routes for the
func (k *Keeper) RegisterRoute(ctx sdk.Context, invarRoute InvarRoute) {
	k.routes = append(k.routes, invarRoute)
}
