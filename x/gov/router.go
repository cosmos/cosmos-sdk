package gov

import (
	"fmt"
	"regexp"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Router interface {
	AddRoute(r string, h sdk.ProposalHandler) (rtr Router)
	Route(path string) (h sdk.ProposalHandler)
}

type router struct {
	routes map[string]sdk.ProposalHandler
}

var _ Router = (*router)(nil)

func NewRouter() *router {
	return &router{
		routes: make(map[string]sdk.ProposalHandler),
	}
}

var isAlphaNumeric = regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString

func (rtr *router) AddRoute(path string, h sdk.ProposalHandler) Router {
	if !isAlphaNumeric(path) {
		panic("route expressions can only contain alphanumeric characters")
	}
	if rtr.routes[path] != nil {
		panic(fmt.Sprintf("route %s has already been initialized", path))
	}

	rtr.routes[path] = h
	return rtr
}

func (rtr *router) Route(path string) sdk.ProposalHandler {
	return rtr.routes[path]
}
