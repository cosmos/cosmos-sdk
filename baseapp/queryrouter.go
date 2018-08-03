package baseapp

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// QueryRouter provides queryables for each query path.
type QueryRouter interface {
	AddRoute(r string, h sdk.Querier) (rtr QueryRouter)
	Route(path string) (h sdk.Querier)
}

// map a transaction type to a handler and an initgenesis function
type queryroute struct {
	r string
	h sdk.Querier
}

type queryrouter struct {
	routes []queryroute
}

// nolint
// NewRouter - create new router
// TODO either make Function unexported or make return type (router) Exported
func NewQueryRouter() *queryrouter {
	return &queryrouter{
		routes: make([]queryroute, 0),
	}
}

// AddRoute - TODO add description
func (rtr *queryrouter) AddRoute(r string, h sdk.Querier) QueryRouter {
	if !isAlphaNumeric(r) {
		panic("route expressions can only contain alphanumeric characters")
	}
	rtr.routes = append(rtr.routes, queryroute{r, h})

	return rtr
}

// Route - TODO add description
// TODO handle expressive matches.
func (rtr *queryrouter) Route(path string) (h sdk.Querier) {
	for _, route := range rtr.routes {
		if route.r == path {
			return route.h
		}
	}
	return nil
}
