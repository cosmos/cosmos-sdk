package types

// Handler defines the core of the state transition function of an application.
// A reference to a Result type must be returned upon success and an error upon
// failure.
type Handler func(ctx Context, msg Msg) (*Result, error)

// AnteHandler authenticates transactions, before their internal messages are handled.
// If newCtx.IsZero(), ctx is used instead.
type AnteHandler func(ctx Context, tx Tx, simulate bool) (newCtx Context, result Result, abort bool)
