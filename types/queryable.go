package types

import (
	abci "github.com/tendermint/tendermint/abci/types"
)

// Querier defines a function type that a module querier must implement to handle
// custom client queries.
type Querier = func(ctx Context, path []string, req abci.RequestQuery) ([]byte, error)
