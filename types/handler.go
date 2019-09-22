package types

// Handler defines the core of the state transition function of an application.
type Handler func(ctx Context, msg Msg) Result

// AnteHandler authenticates transactions, before their internal messages are handled.
// If newCtx.IsZero(), ctx is used instead.
type AnteHandler func(ctx Context, tx Tx, simulate bool) (newCtx Context, err error)

// AnteDecorator wraps the next AnteHandler to perform custom pre- and post-processing.
type AnteDecorator interface {
	AnteHandle(ctx Context, tx Tx, simulate bool, next AnteHandler) (newCtx Context, err error)
}

// ChainDecorator chains AnteDecorators together with each element
// wrapping over the decorators further along chain and returns a single AnteHandler.
//
// First element is outermost decorator, last element is innermost decorator
func ChainDecorators(chain ...AnteDecorator) AnteHandler {
	chain = append(chain, Tail{})
	if len(chain) == 1 {
		return func(ctx Context, tx Tx, simulate bool) (Context, error) {
			return chain[0].AnteHandle(ctx, tx, simulate, nil)
		}
	}
	return func(ctx Context, tx Tx, simulate bool) (Context, error) {
		return chain[0].AnteHandle(ctx, tx, simulate, ChainDecorators(chain[1:]...))
	}
}

// Tail AnteDecorator will get added to the chain to simplify decorator code
// Don't need to check if next == nil further up the chain
type Tail struct{}

// Simply return provided Context and nil error
func (t Tail) AnteHandle(ctx Context, tx Tx, simulate bool, next AnteHandler) (Context, error) {
	return ctx, nil
}
