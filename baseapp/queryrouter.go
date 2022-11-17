package baseapp

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type QueryRouter struct {
	routes map[string]sdk.Querier
}

var _ sdk.QueryRouter = NewQueryRouter()

// NewQueryRouter returns a reference to a new QueryRouter.
func NewQueryRouter() *QueryRouter {
	return &QueryRouter{
		routes: map[string]sdk.Querier{},
	}
}

// AddRoute adds a query path to the router with a given Querier. It will panic
// if a duplicate route is given. The route must be alphanumeric.
func (qrt *QueryRouter) AddRoute(route string, q sdk.Querier) sdk.QueryRouter {
	if !sdk.IsAlphaNumeric(route) {
		panic("route expressions can only contain alphanumeric characters")
	}

	// paths are only the final extensions!
	// Needed to ensure erroneous queries don't get into the state machine.
	if strings.Contains(route, "/") {
		panic("route's don't contain '/'")
	}

	if qrt.routes[route] != nil {
		panic(fmt.Sprintf("route %s has already been initialized", route))
	}

	qrt.routes[route] = q

	return qrt
}

// Route returns the Querier for a given query route path.
func (qrt *QueryRouter) Route(path string) sdk.Querier {
	return qrt.routes[path]
}
