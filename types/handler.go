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
func ChainAnteDecorators(chain ...AnteDecorator) AnteHandler {
	if (chain[len(chain)-1] != Terminator{}) {
		chain = append(chain, Terminator{})
	}
	if len(chain) == 1 {
		return func(ctx Context, tx Tx, simulate bool) (Context, error) {
			return chain[0].AnteHandle(ctx, tx, simulate, nil)
		}
	}
	return func(ctx Context, tx Tx, simulate bool) (Context, error) {
		return chain[0].AnteHandle(ctx, tx, simulate, ChainAnteDecorators(chain[1:]...))
	}
}

// Terminator AnteDecorator will get added to the chain to simplify decorator code
// Don't need to check if next == nil further up the chain
//                        ______
//                     <((((((\\\
//                     /      . }\
//                     ;--..--._|}
//  (\                 '--/\--'  )
//   \\                | '-'  :'|
//    \\               . -==- .-|
//     \\               \.__.'   \--._
//     [\\          __.--|       //  _/'--.
//     \ \\       .'-._ ('-----'/ __/      \
//      \ \\     /   __>|      | '--.       |
//       \ \\   |   \   |     /    /       /
//        \ '\ /     \  |     |  _/       /
//         \  \       \ |     | /        /
//   snd    \  \      \        /
type Terminator struct{}

// Simply return provided Context and nil error
func (t Terminator) AnteHandle(ctx Context, tx Tx, simulate bool, next AnteHandler) (Context, error) {
	return ctx, nil
}
