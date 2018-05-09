package sdk

import abci "github.com/tendermint/abci/types"

// initialize application state at genesis
type InitChainer func(ctx Context, req abci.RequestInitChain) abci.ResponseInitChain

// run code before the transactions in a block
type BeginBlocker func(ctx Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock

// run code after the transactions in a block and return updates to the validator set
type EndBlocker func(ctx Context, req abci.RequestEndBlock) abci.ResponseEndBlock

// runs to determine whether to add tx to the mempool
type CheckTxer func(ctx Context, txBytes []byte) Result

// runs to execute tx in block
type DeliverTxer func(ctx Context, txBytes []byte) Result
