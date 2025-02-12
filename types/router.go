package types

import (
	"regexp"

	abci "github.com/cometbft/cometbft/abci/types"
)

// IsAlphaNumeric defines a regular expression for matching against alpha-numeric
// values.
var IsAlphaNumeric = regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString

// Querier defines a function type that a module querier must implement to handle
// custom client queries.
type Querier = func(ctx Context, path []string, req *abci.RequestQuery) ([]byte, error)

// QueryRouter provides queryables for each query path.
type QueryRouter interface {
	AddRoute(r string, h Querier) QueryRouter
	Route(path string) Querier
}
