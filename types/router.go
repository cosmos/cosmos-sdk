package types

// Router provides handlers for each transaction type.
type Router interface {
	AddRoute(r string, h Handler) Router
	Route(path string) Handler
}

// QueryRouter provides queryables for each query path.
type QueryRouter interface {
	AddRoute(r string, h Querier) QueryRouter
	Route(path string) Querier
}
