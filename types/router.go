package types

import "regexp"

// IsAlphaNumeric defines a regular expression for matching against alpha-numeric
// values.
var IsAlphaNumeric = regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString

// Router provides handlers for each transaction type.
type Router interface {
	AddRoute(r string, h Handler) Router
	Route(ctx Context, path string) Handler
}

// QueryRouter provides queryables for each query path.
type QueryRouter interface {
	AddRoute(r string, h Querier) QueryRouter
	Route(path string) Querier
}
