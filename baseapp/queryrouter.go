package baseapp

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// QueryRouter provides queryables for each query path.
type QueryRouter interface {
	AddRoute(r string, h sdk.Querier) (rtr QueryRouter)
	Route(path string) (h sdk.Querier)
}

type queryRouter struct {
	routes map[string]sdk.Querier
}

// NewQueryRouter returns a reference to a new queryRouter.
//
// TODO: Either make the function private or make return type (queryRouter) public.
func NewQueryRouter() *queryRouter { // nolint: golint
	return &queryRouter{
		routes: map[string]sdk.Querier{},
	}
}

// AddRoute adds a query path to the router with a given Querier. It will panic
// if a duplicate route is given. The route must be alphanumeric.
func (qrt *queryRouter) AddRoute(r string, q sdk.Querier) QueryRouter {
	if !isAlphaNumeric(r) {
		panic("route expressions can only contain alphanumeric characters")
	}
	if qrt.routes[r] != nil {
		panic("route has already been initialized")
	}

	qrt.routes[r] = q
	return qrt
}

// Route returns the Querier for a given query route path.
func (qrt *queryRouter) Route(path string) (h sdk.Querier) {
	return qrt.routes[path]
}
