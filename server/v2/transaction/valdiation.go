package transaction

import (
	"context"

	tx "github.com/cosmos/cosmos-sdk/serverv2/core/transaction"
)

var _ tx.Validator[tx.Tx] = TxValidator[tx.Tx]{}

// TxValidator is a transaction validator that validates transactions based off an existing set of handlers
// TxValidator is designed to be synchronous
type TxValidator[T tx.Tx] struct {
	handler tx.Handler
}

func NewTxValidator[T tx.Tx]() *TxValidator[T] {
	return &TxValidator[T]{}
}

// RegisterHandler registers the handlers to the transaction verifier.
// The order of the handlers is important. The order passed here is the order of execution
func (v *TxValidator[T]) RegisterHandler(h tx.Handler) {
	v.handler = h
}

// Validate validates the transaction
// it returns the context to be used further for execution
func (v TxValidator[T]) Validate(ctx context.Context, txs []T, simulate bool) (context.Context, map[[32]byte]error) {
	var (
		errMap = make(map[[32]byte]error, len(txs)) // used to return a map of which txs failed
		newctx = ctx                                // retrun the context to be used further for execution
	)

	for _, tx := range txs {
		// create a copy of the context for each transaction
		cctx := newctx
		ctx, err := v.handler(cctx, tx, simulate)
		if err != nil {
			errMap[tx.Hash()] = err
		} else {
			// if no error, set the context on the newctx variable
			newctx = ctx
		}

	}

	return newctx, errMap
}
