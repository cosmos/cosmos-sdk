package types

import abci "github.com/tendermint/abci/types"

// initialize application state at genesis
type InitChainer func(ctx Context, req abci.RequestInitChain) abci.ResponseInitChain

//
type BeginBlocker func(ctx Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock

//
type EndBlocker func(ctx Context, req abci.RequestEndBlock) abci.ResponseEndBlock
