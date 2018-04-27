package types

// core function variable which application runs for transactions
type Handler func(ctx Context, msg Msg) Result

// core function variable which application runs to handle fees
type FeeHandler func(ctx Context, tx Tx, fee Coins)

// If newCtx.IsZero(), ctx is used instead.
type AnteHandler func(ctx Context, tx Tx) (newCtx Context, result Result, abort bool)
