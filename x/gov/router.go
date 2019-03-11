package gov

import (
	"fmt"
	"regexp"

	"github.com/cosmos/cosmos-sdk/x/gov/proposal"
)

type Router interface {
	AddRoute(r string, h proposal.Handler) (rtr Router)
	Route(path string) (h proposal.Handler)
}

type router struct {
	routes map[string]proposal.Handler
}

var _ Router = (*router)(nil)

func NewRouter() *router {
	return &router{
		routes: make(map[string]proposal.Handler),
	}
}

var isAlphaNumeric = regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString

func (rtr *router) AddRoute(path string, h proposal.Handler) Router {
	if !isAlphaNumeric(path) {
		panic("route expressions can only contain alphanumeric characters")
	}
	if rtr.routes[path] != nil {
		panic(fmt.Sprintf("route %s has already been initialized", path))
	}

	rtr.routes[path] = h
	return rtr
}

func (rtr *router) Route(path string) proposal.Handler {
	return rtr.routes[path]
}
