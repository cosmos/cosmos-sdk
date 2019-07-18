package channel

type Router interface {
	AddRoute(path string, h Handler) Router
	Route(path string) Handler
}

type router struct {
	routes map[string]Handler
}

func NewRouter() Router {
	return &router{
		routes: make(map[string]Handler),
	}
}

func (router *router) AddRoute(path string, h Handler) Router {
	// TODO
	/*
		if !isAlphaNumeric(path) {
			panic("route expressions can only contain alphanumeric characters")
		}
	*/
	if router.routes[path] != nil {
		panic("route " + path + "has already been initialized")
	}
	router.routes[path] = h
	return router
}

func (router *router) Route(path string) Handler {
	return router.routes[path]
}
