package types

type Handler func(ctx Context, tx Tx) Result

type AnteHandler func(ctx Context, tx Tx) (Result, abort bool)
