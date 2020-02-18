package types

import "regexp"

// IsAlphaNumeric defines a regular expression for matching against alpha-numeric
// values.
var IsAlphaNumeric = regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString

// IsAlphaLower defines regular expression to check if the string has lowercase
// alphabetic characters only.
var IsAlphaLower = regexp.MustCompile(`^[a-z]+$`).MatchString

// IsAlphaUpper defines regular expression to check if the string has uppercase
// alphabetic characters only.
var IsAlphaUpper = regexp.MustCompile(`^[A-Z]+$`).MatchString

// IsAlpha defines regular expression to check if the string has alphabetic
// characters only.
var IsAlpha = regexp.MustCompile(`^[a-zA-Z]+$`).MatchString

// IsNumeric defines regular expression to check if the string has numeric
// characters only.
var IsNumeric = regexp.MustCompile(`^[0-9]+$`).MatchString

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
