package baseapp

import (
	"regexp"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Router provides handlers for each transaction type.
type Router interface {
	AddRoute(h sdk.Handler) (rtr Router)
	Route(path string) (h sdk.Handler)
}

// map a transaction type to a handler and an initgenesis function

type router struct {
	handlers []sdk.Handler
}

// nolint
// NewRouter - create new router
// TODO either make Function unexported or make return type (router) Exported
func NewRouter() *router {
	return &router{
		handlers: make([]sdk.Handler, 0),
	}
}

var isAlpha = regexp.MustCompile(`^[a-zA-Z]+$`).MatchString

// AddRoute - TODO add description
func (rtr *router) AddRoute(h sdk.Handler) Router {
	if !isAlpha(h.Type()) {
		panic("route expressions can only contain alphanumeric characters")
	}
	rtr.handlers = append(rtr.handlers, h)

	return rtr
}

// Route - TODO add description
// TODO handle expressive matches.
func (rtr *router) Route(path string) (h sdk.Handler) {
	for _, h := range rtr.handlers {
		if h.Type() == path {
			return h
		}
	}
	return nil
}
