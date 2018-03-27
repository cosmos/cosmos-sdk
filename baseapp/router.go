package baseapp

import (
	"regexp"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Router provides handlers for each transaction type.
type Router interface {
	AddRoute(r string, h sdk.Handler, i sdk.InitGenesis) (rtr Router)
	Route(path string) (h sdk.Handler)
	RouteGenesis(path string) (i sdk.InitGenesis)
	ForEach(func(r string, h sdk.Handler, i sdk.InitGenesis) error) error
}

// map a transaction type to a handler and an initgenesis function
type route struct {
	r string
	h sdk.Handler
	i sdk.InitGenesis
}

type router struct {
	routes []route
}

// nolint
// NewRouter - create new router
// TODO either make Function unexported or make return type (router) Exported
func NewRouter() *router {
	return &router{
		routes: make([]route, 0),
	}
}

var isAlpha = regexp.MustCompile(`^[a-zA-Z]+$`).MatchString

// AddRoute - TODO add description
func (rtr *router) AddRoute(r string, h sdk.Handler, i sdk.InitGenesis) Router {
	if !isAlpha(r) {
		panic("route expressions can only contain alphanumeric characters")
	}
	rtr.routes = append(rtr.routes, route{r, h, i})

	return rtr
}

// TODO handle expressive matches.
func matchRoute(path string, route string) bool {
	return path == route
}

// Route - TODO add description
func (rtr *router) Route(path string) (h sdk.Handler) {
	for _, route := range rtr.routes {
		if matchRoute(path, route.r) {
			return route.h
		}
	}
	return nil
}

func (rtr *router) RouteGenesis(path string) (i sdk.InitGenesis) {
	for _, route := range rtr.routes {
		if matchRoute(path, route.r) {
			return route.i
		}
	}
	return nil
}

func (rtr *router) ForEach(f func(string, sdk.Handler, sdk.InitGenesis) error) error {
	for _, route := range rtr.routes {
		if err := f(route.r, route.h, route.i); err != nil {
			return err
		}
	}
	return nil
}
