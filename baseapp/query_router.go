package baseapp

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.QueryRouter = NewQueryRouter()

// queryRouterCustom implements the QueryRouter interface.
type queryRouterCustom struct {
	routes map[string]sdk.Querier
}

// NewQueryRouter returns a reference to a new QueryRouter. This should be used for register custom abci
// query handlers on the base application.
func NewQueryRouter() sdk.QueryRouter {
	return &queryRouterCustom{
		routes: map[string]sdk.Querier{},
	}
}

// AddRoute adds a query path to the router with a given Querier. It will panic
// if a duplicate route is given. The route must be alphanumeric.
func (qrt *queryRouterCustom) AddRoute(route string, q sdk.Querier) sdk.QueryRouter {
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
func (qrt *queryRouterCustom) Route(path string) sdk.Querier {
	return qrt.routes[path]
}
