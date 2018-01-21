package baseapp

import (
	"regexp"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Router interface {
	AddRoute(r string, h sdk.Handler)
	Route(path string) (h sdk.Handler)
}

type route struct {
	r string
	h sdk.Handler
}

type router struct {
	routes []route
}

func NewRouter() *router {
	return &router{
		routes: make([]route, 0),
	}
}

var isAlpha = regexp.MustCompile(`^[a-zA-Z]+$`).MatchString

func (rtr *router) AddRoute(r string, h sdk.Handler) {
	if !isAlpha(r) {
		panic("route expressions can only contain alphanumeric characters")
	}
	rtr.routes = append(rtr.routes, route{r, h})
}

// TODO handle expressive matches.
func (rtr *router) Route(path string) (h sdk.Handler) {
	for _, route := range rtr.routes {
		if route.r == path {
			return route.h
		}
	}
	return nil
}
