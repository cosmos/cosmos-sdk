package baseapp

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// QueryRouter provides queryables for each query path.
type QueryRouter interface {
	AddRoute(r string, h sdk.Querier) (rtr QueryRouter)
	Route(path string) (h sdk.Querier)
}

type queryrouter struct {
	routes map[string]sdk.Querier
}

// nolint
// NewRouter - create new router
// TODO either make Function unexported or make return type (router) Exported
func NewQueryRouter() *queryrouter {
	return &queryrouter{
		routes: map[string]sdk.Querier{},
	}
}

// AddRoute - Adds an sdk.Querier to the route provided. Panics on duplicate
func (rtr *queryrouter) AddRoute(r string, q sdk.Querier) QueryRouter {
	if !isAlphaNumeric(r) {
		panic("route expressions can only contain alphanumeric characters")
	}
	if rtr.routes[r] != nil {
		panic("route has already been initialized")
	}
	rtr.routes[r] = q
	return rtr
}

// Returns the sdk.Querier for a certain route path
func (rtr *queryrouter) Route(path string) (h sdk.Querier) {
	return rtr.routes[path]
}
