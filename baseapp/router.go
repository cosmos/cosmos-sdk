package baseapp

import (
	"regexp"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Router provides handlers for each transaction type.
type Router interface {
	AddRoute(r string, h sdk.Handler) (rtr Router)
	Route(path string) (h sdk.Handler)
}

// map a transaction type to a handler and an initgenesis function
type route struct {
	r string
	h sdk.Handler
}

type router struct {
	routes []route
}

// NewRouter returns a reference to a new router.
//
// TODO: Either make the function private or make return type (router) public.
func NewRouter() *router { // nolint: golint
	return &router{
		routes: make([]route, 0),
	}
}

var isAlphaNumeric = regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString

// AddRoute adds a route path to the router with a given handler. The route must
// be alphanumeric.
func (rtr *router) AddRoute(r string, h sdk.Handler) Router {
	if !isAlphaNumeric(r) {
		panic("route expressions can only contain alphanumeric characters")
	}

	rtr.routes = append(rtr.routes, route{r, h})
	return rtr
}

// Route returns a handler for a given route path.
//
// TODO: Handle expressive matches.
func (rtr *router) Route(path string) (h sdk.Handler) {
	for _, route := range rtr.routes {
		if route.r == path {
			return route.h
		}
	}
	return nil
}
