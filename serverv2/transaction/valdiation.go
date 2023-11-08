package transaction

import (
	"context"

	tx "cosmossdk.io/core/transaction"
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
func (v TxValidator[T]) RegisterHandler(h tx.Handler) {
	v.handler = h
}

// Validate validates the transaction
// it returns the context to be used further for execution
func (v TxValidator[T]) Validate(ctx context.Context, txs []T) (context.Context, error) {
	for _, tx := range txs {
		ctx, err := v.handler(ctx, tx, simulate)
		return ctx, err
	}

	return nil, nil
}
