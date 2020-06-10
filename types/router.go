package types

import "regexp"

var (
	// IsAlphaNumeric defines a regular expression for matching against alpha-numeric
	// values.
	IsAlphaNumeric = regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString

	// IsAlphaLower defines regular expression to check if the string has lowercase
	// alphabetic characters only.
	IsAlphaLower = regexp.MustCompile(`^[a-z]+$`).MatchString

	// IsAlphaUpper defines regular expression to check if the string has uppercase
	// alphabetic characters only.
	IsAlphaUpper = regexp.MustCompile(`^[A-Z]+$`).MatchString

	// IsAlpha defines regular expression to check if the string has alphabetic
	// characters only.
	IsAlpha = regexp.MustCompile(`^[a-zA-Z]+$`).MatchString

	// IsNumeric defines regular expression to check if the string has numeric
	// characters only.
	IsNumeric = regexp.MustCompile(`^[0-9]+$`).MatchString
)

// Router provides handlers for each transaction type.
type Router interface {
	AddRoute(r Route) Router
	Route(ctx Context, path string) Handler
}

// Route represents a Route used by the Router.
type Route interface {
	Path() string
	Handler() Handler
}

type route struct {
	route   string
	handler Handler
}

func NewRoute(r string, h Handler) Route {
	return &route{route: r, handler: h}
}

func (r route) Path() string {
	return r.route
}

func (r route) Handler() Handler {
	return r.handler
}

// QueryRouter provides queryables for each query path.
type QueryRouter interface {
	AddRoute(r string, h Querier) QueryRouter
	Route(path string) Querier
}
