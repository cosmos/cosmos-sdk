package types

type Handler func(ctx Context, tx Tx) Result

type AnteHandler func(ctx Context, tx Tx) (newCtx Context, result Result, abort bool)
