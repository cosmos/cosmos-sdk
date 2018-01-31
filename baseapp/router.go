package baseapp

import (
	"regexp"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Router interface {
	AddRoute(r string, h sdk.Handler)
	AddAnte(r string, a sdk.AnteHandler)
	Route(path string) (h sdk.Handler)
	Ante(path string) (a sdk.AnteHandler)
}

type route struct {
	r string
	h sdk.Handler
	a sdk.AnteHandler
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
	rtr.routes = append(rtr.routes, route{r, h, nil})
}

func (rtr *router) AddAnte(r string, a sdk.AnteHandler) {
	if !isAlpha(r) {
		panic("route expressions can only contain alphanumeric characters")
	}

	for _, route := range rtr.routes {
		if route.r == r {
			route.a = a;
			return;
		}
	}

	panic("ante can only be added after routes")

}

func (rtr *router) Route(path string) (h sdk.Handler) {
	for _, route := range rtr.routes {
		if route.r == path {
			return route.h
		}
	}
	return nil
}

func (rtr *router) Ante(path string) (a sdk.AnteHandler) {
	for _, route := range rtr.routes {
		if route.r == path {
			return route.a
		}
	}
	return nil
}
