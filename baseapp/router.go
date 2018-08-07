package baseapp

import (
	"regexp"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"strings"
)

// Router provides handlers for each transaction type.
type Router interface {

	////////////////////  iris/cosmos-sdk begin  ///////////////////////////
	AddRoute(r string, s []*sdk.KVStoreKey, h sdk.Handler) (rtr Router)
	Route(path string) (h sdk.Handler)
	RouteTable() (table []string)
	////////////////////  iris/cosmos-sdk end  ///////////////////////////
}

// map a transaction type to a handler and an initgenesis function
type route struct {
	r string
	////////////////////  iris/cosmos-sdk begin  ///////////////////////////
	s []*sdk.KVStoreKey
	////////////////////  iris/cosmos-sdk end  ///////////////////////////
	h sdk.Handler
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
////////////////////  iris/cosmos-sdk begin  ///////////////////////////
func (rtr *router) AddRoute(r string, s []*sdk.KVStoreKey, h sdk.Handler) Router {
	rstrs := strings.Split(r, "-")

	if !isAlpha(rstrs[0]) {
		panic("route expressions can only contain alphabet characters")
	}
	rtr.routes = append(rtr.routes, route{r, s, h})

	return rtr
}

////////////////////  iris/cosmos-sdk end  ///////////////////////////

// Route - TODO add description
// TODO handle expressive matches.
func (rtr *router) Route(path string) (h sdk.Handler) {
	for _, route := range rtr.routes {
		if route.r == path {
			return route.h
		}
	}
	return nil
}

////////////////////  iris/cosmos-sdk begin  ///////////////////////////

func (rtr *router) RouteTable() (table []string) {
	for _, route := range rtr.routes {
		storelist := ""
		for _, store := range route.s {
			if storelist == "" {
				storelist = store.Name()
			} else {
				storelist = storelist + ":" + store.Name()
			}
		}
		table = append(table, route.r+"/"+storelist)
	}
	return
}

////////////////////  iris/cosmos-sdk end  ///////////////////////////
