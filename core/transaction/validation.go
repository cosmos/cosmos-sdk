package transaction

import "context"

type Tx interface {
	Hash() [32]byte // TODO evaluate if 32 bytes is the right size
}

// Handler is a recursive function that takes a transaction and returns a new context to be used by the next handler
type Handler func(ctx context.Context, tx Tx, simulate bool) (newCtx context.Context, err error)

// Validator is a transaction validator that validates transactions based off an existing set of handlers
// Validators can be designed to be asynchronous or synchronous
type Validator[T Tx] interface {
	Validate(context.Context, []T, bool) (context.Context, error)
}
