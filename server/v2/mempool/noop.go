package mempool

import (
	"context"

	"cosmossdk.io/server/v2/core/mempool"
	"cosmossdk.io/server/v2/core/transaction"
)

var _ mempool.Mempool[transaction.Tx] = NoOpMempool[transaction.Tx]{}

type NoOpMempool[T transaction.Tx] struct {
	txValidator transaction.Validator[T]
}

func NewNoopMempool[T transaction.Tx](txv transaction.Validator[T]) *NoOpMempool[T] {
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

func (npm NoOpMempool[T]) Insert(ctx context.Context, txs []T) map[[32]byte]error {
	_, err := npm.txValidator.Validate(ctx, txs)
	return err
}

func (NoOpMempool[T]) GetTxs(ctx context.Context, size uint32) ([]T, error) {
	return []T{}, nil
}

func (NoOpMempool[T]) CountTx() uint32 { return 0 }
func (NoOpMempool[T]) Remove(T) error  { return nil }
