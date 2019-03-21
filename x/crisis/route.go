package crisis

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// invariant route
type InvarRoute struct {
	Route string
	Invar sdk.Invariant
}

// NewInvarRoute - create an InvarRoute object
func NewInvarRoute(route string, invar sdk.Invariant) InvarRoute {
	return InvarRoute{
		Route: route,
		Invar: invar,
	}
}
