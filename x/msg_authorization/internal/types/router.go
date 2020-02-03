package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ Router = (*router)(nil)

type Router interface {
	AddRoute(r string, h sdk.Handler) (rtr Router)
	Seal()
	Route(c sdk.Context, p string) (h sdk.Handler)
}

type router struct {
	routes map[string]sdk.Handler
	sealed bool
}

// NewRouter creates a new Router interface instance
func NewRouter() Router {
	return &router{
		routes: make(map[string]sdk.Handler),
	}
}

// Seal seals the router which prohibits any subsequent route handlers to be
// added. Seal will panic if called more than once.
func (rtr *router) Seal() {
	if rtr.sealed {
		panic("router already sealed")
	}
	rtr.sealed = true
}

// AddRoute adds a governance handler for a given path. It returns the Router
// so AddRoute calls can be linked. It will panic if the router is sealed.
func (rtr *router) AddRoute(path string, h sdk.Handler) Router {
	if rtr.sealed {
		panic("router sealed; cannot add route handler")
	}

	if !sdk.IsAlphaNumeric(path) {
		panic("route expressions can only contain alphanumeric characters")
	}

	rtr.routes[path] = h
	return rtr
}

func (rtr *router) Route(_ sdk.Context, path string) sdk.Handler {
	return rtr.routes[path]
}
