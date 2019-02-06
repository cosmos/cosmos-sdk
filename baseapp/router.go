package baseapp

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Router provides handlers for each transaction type.
type Router interface {
	AddRoute(r string, h sdk.Handler) (rtr Router)
	Route(path string) (h sdk.Handler)
}

type router struct {
	routes map[string]sdk.Handler
}

// NewRouter returns a reference to a new router.
//
// TODO: Either make the function private or make return type (router) public.
func NewRouter() *router { // nolint: golint
	return &router{
		routes: make(map[string]sdk.Handler),
	}
}

// AddRoute adds a route path to the router with a given handler. The route must
// be alphanumeric.
func (rtr *router) AddRoute(path string, h sdk.Handler) Router {
	if !isAlphaNumeric(path) {
		panic("route expressions can only contain alphanumeric characters")
	}
	if rtr.routes[path] != nil {
		panic(fmt.Sprintf("route %s has already been initialized", path))
	}

	rtr.routes[path] = h
	return rtr
}

// Route returns a handler for a given route path.
//
// TODO: Handle expressive matches.
func (rtr *router) Route(path string) sdk.Handler {
	return rtr.routes[path]
}
