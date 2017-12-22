package types

// Handler handles both ABCI DeliverTx and CheckTx requests.
// Iff ABCI.CheckTx, ctx.IsCheckTx() returns true.
type Handler func(ctx Context, ms MultiStore, tx Tx) Result
