package crisis

import (
	"fmt"

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

// InvarRoutes an array of InvarRoute
type InvarRoutes []InvarRoute

// Get an array of just the routes
func (irs InvarRoutes) Routes() Routes {
	var routes []string
	for _, ir := range irs {
		routes = append(routes, ir.Route)
	}
	return routes
}

// routes
type Routes []string

func (rs Routes) String() string {
	var out string
	for _, r := range rs {
		out += fmt.Sprintf("%v\n", r)
	}
	return out
}
