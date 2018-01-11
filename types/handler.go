package types

// Handler handles both ABCI DeliverTx and CheckTx requests.
// Iff ABCI.CheckTx, ctx.IsCheckTx() returns true.
type Handler func(ctx Context, tx Tx) Result
