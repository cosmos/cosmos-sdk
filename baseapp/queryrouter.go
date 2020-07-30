package baseapp

import (
	"fmt"

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
func (qrt *QueryRouter) AddRoute(path string, q sdk.Querier) sdk.QueryRouter {
	if !sdk.IsAlphaNumeric(path) {
		panic("route expressions can only contain alphanumeric characters")
	}

	if qrt.routes[path] != nil {
		panic(fmt.Sprintf("route %s has already been initialized", path))
	}

	qrt.routes[path] = q

	return qrt
}

// Route returns the Querier for a given query route path.
func (qrt *QueryRouter) Route(path string) sdk.Querier {
	return qrt.routes[path]
}
