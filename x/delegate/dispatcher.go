package delegate

import (
	cosmos "github.com/cosmos/cosmos-sdk/types"
	"regexp"
)

type dispatcher struct {
	Keeper
	Router
}

func NewDispatcher(k Keeper, r Router) Dispatcher {
	return &dispatcher{k, r}
}

func (dispatcher dispatcher) DispatchAction(ctx cosmos.Context, sender cosmos.AccAddress, action Action) cosmos.Result {
	caps := action.RequiredCapabilities()
	for _, c := range caps {
		if !dispatcher.HasCapability(ctx, sender, ActorCapability{c, action.Actor()}) {
			return cosmos.ErrUnauthorized("actor does not have capability").Result()
		}
	}
	for _, route := range dispatcher.routes {
		if route.r == action.Route() {
			return route.h.HandleAction(ctx, action)
		}
	}
	return cosmos.ErrUnknownRequest("can't find action handler").Result()
}

type ActionHandler interface {
	HandleAction(ctx cosmos.Context, action Action) cosmos.Result
}
type route struct {
	r string
	h ActionHandler
}

type Router struct {
	// TODO change this to a map
	routes []route
}

func NewRouter() *Router {
	return &Router{
		routes: make([]route, 0),
	}
}

var isAlphaNumeric = regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString

func (rtr *Router) AddRoute(r string, h ActionHandler) *Router {
	if !isAlphaNumeric(r) {
		panic("route expressions can only contain alphanumeric characters")
	}
	rtr.routes = append(rtr.routes, route{r, h})

	return rtr
}
