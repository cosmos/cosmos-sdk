package types

type Handler func(ctx Context, tx Tx) Result

// If newCtx.IsZero(), ctx is used instead.
type AnteHandler func(ctx Context, tx Tx) (newCtx Context, result Result, abort bool)
