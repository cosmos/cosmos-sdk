package types

import abci "github.com/tendermint/abci/types"

// initialize application state at genesis
type InitChainer func(ctx Context, req abci.RequestInitChain) Error
