package mempool

import (
	"context"

	"cosmossdk.io/server/v2/core/mempool"
	"cosmossdk.io/server/v2/core/transaction"
)

var _ mempool.Mempool[transaction.Tx] = NoOpMempool[transaction.Tx]{}

type NoOpMempool[T transaction.Tx] struct{}

func NewNoopMempool[T transaction.Tx]() NoOpMempool[T] { return NoOpMempool[T]{} }

func (m NoOpMempool[T]) Start() error {
	// NoOpMempool[T] does not require any initialization
	return nil
}

func (m NoOpMempool[T]) Stop() error {
	// NoOpMempool[T] does not require any cleanup
	return nil
}

func (m NoOpMempool[T]) Insert(ctx context.Context, tx T) error { return nil }

func (NoOpMempool[T]) Get(ctx context.Context, size int) ([]T, error) { return nil, nil }

func (NoOpMempool[T]) Remove(txs []T) error { return nil }
