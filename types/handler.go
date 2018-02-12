package types

type Handler func(ctx Context, msg Msg) Result

// AnteHandler is an action that is run no matter whether the transaction
// succeeds or fails. It should check and do things that are applicable to all
// transactions. An Ethereum example is to deduct the fees.
// If newCtx.IsZero(), ctx is used instead.
type AnteHandler func(ctx Context, tx Tx) (newCtx Context, result Result, abort bool)
