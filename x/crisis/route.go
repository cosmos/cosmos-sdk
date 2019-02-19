package crisis

import sdk "github.com/cosmos/cosmos-sdk/types"

// invariant route
type InvarRoute struct {
	Route string
	Invar sdk.Invariant
}

// InvarRoute - TODO
func NewInvarRoute(route string, invar sdk.Invariant) InvarRoute {
	return InvarRoute{
		Route: route,
		Invar: invar,
	}
}

type InvarRoutes []InvarRoute

func (i *InvarRoutes) RegisterInvar(route string, invar sdk.Invariant) {
	invarRoute := NewInvarRoute(route, invar)
	i = append(i, invarRoute)
}
