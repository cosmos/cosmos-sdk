package app

type Router interface {
	AddRoute(r string, h Handler)
	Route(path string) (h Handler)
}

type route struct {
	r string
	h Handler
}

type router struct {
	routes []route
}

func NewRouter() router {
	return router{
		routes: make([]route),
	}
}

var isAlpha = regexp.MustCompile(`^[a-zA-Z]+$`).MatchString

func (rtr router) AddRoute(r string, h Handler) {
	if !isAlpha(r) {
		panic("route expressions can only contain alphanumeric characters")
	}
	rtr.routes = append(rtr.routes, route{r, h})
}

// TODO handle expressive matches.
func (rtr router) Route(path string) (h Handler) {
	for _, route := range rtr.routes {
		if route.r == path {
			return route.h
		}
	}
	return nil
}
