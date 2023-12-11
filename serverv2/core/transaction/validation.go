package transaction

import "context"

type Tx interface {
	Hash() [32]byte // TODO evaluate if 32 bytes is the right size & benchmark overhead of hashing instead of using identifier
}

// Handler is a recursive function that takes a transaction and returns a new context to be used by the next handler
type Handler func(ctx context.Context, tx Tx) (newCtx context.Context, err error)

// Validator is a transaction validator that validates transactions based off an existing set of handlers
// Validators can be designed to be asynchronous or synchronous
type Validator[T Tx] interface {
	// Validate validates the transactions
	// it returns the context used and a map of which txs failed.
	// It does not take into account what information is needed to be returned to the consensus engine, this must be extracted from teh context
	Validate(ctx context.Context, txs []T) (context.Context, map[[32]byte]error)
}
