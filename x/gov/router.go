package gov

import (
	"fmt"
	"regexp"

	"github.com/cosmos/cosmos-sdk/x/gov/proposal"
)

// Router is a map from string to proposal.Handler
// copied and modified from baseapp/router.go
type Router interface {
	AddRoute(r string, h proposal.Handler) (rtr Router)
	HasRoute(r string) bool
	GetRoute(path string) (h proposal.Handler)
}

type router struct {
	routes map[string]proposal.Handler
}

var _ Router = (*router)(nil)

// Constructs new router
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

func (rtr *router) HasRoute(path string) bool {
	return rtr.routes[path] != nil
}

func (rtr *router) GetRoute(path string) proposal.Handler {
	return rtr.routes[path]
}
