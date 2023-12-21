package mempool

import (
	"context"
)

// TODO here until we rebase to get the txcodec items
type TxValidator[T any] interface {
	ValidateTx(ctx context.Context, tx T, simulate bool) (context.Context, error)
}

var _ Mempool[any] = NoOpMempool[any]{}

type NoOpMempool[T any] struct {
	txValidator TxValidator[T]
}

func NewNoopMempool[T any](txv TxValidator[T]) *NoOpMempool[T] {
	return &NoOpMempool[T]{txValidator: txv}
}

func (s *NoOpMempool[T]) Start() error {
	// NoOpMempool[T] does not require any initialization
	return nil
}

func (s *NoOpMempool[T]) Stop() error {
	// NoOpMempool[T] does not require any cleanup
	return nil
}

func (npm NoOpMempool[T]) Insert(ctx context.Context, tx T) error {
	_, err := npm.txValidator.ValidateTx(ctx, tx, false)
	return err
}

func (NoOpMempool[T]) GetTxs(ctx context.Context, size uint32) (any, error) {
	return nil, nil
}

func (NoOpMempool[T]) CountTx() uint32  { return 0 }
func (NoOpMempool[T]) Remove(any) error { return nil }
