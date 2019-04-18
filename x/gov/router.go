package gov

import (
	"fmt"
	"regexp"
)

var (
	_ Router = (*router)(nil)

	isAlphaNumeric = regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString
)

// Router implements a governance Handler router.
//
// TODO: Use generic router (ref #3976).
type Router interface {
	AddRoute(r string, h Handler) (rtr Router)
	HasRoute(r string) bool
	GetRoute(path string) (h Handler)
}

type router struct {
	routes map[string]Handler
}

func NewRouter() Router {
	return &router{
		routes: make(map[string]Handler),
	}
}

func (rtr *router) AddRoute(path string, h Handler) Router {
	if !isAlphaNumeric(path) {
		panic("route expressions can only contain alphanumeric characters")
	}
	if rtr.HasRoute(path) {
		panic(fmt.Sprintf("route %s has already been initialized", path))
	}

	rtr.routes[path] = h
	return rtr
}

// HasRoute returns true if the router has a path registered or false otherwise.
func (rtr *router) HasRoute(path string) bool {
	return rtr.routes[path] != nil
}

// GetRoute returns a Handler for a given path.
func (rtr *router) GetRoute(path string) Handler {
	if !rtr.HasRoute(path) {
		panic(fmt.Sprintf("route \"%s\" does not exist", path))
	}

	return rtr.routes[path]
}
