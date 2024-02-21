package mempool

import (
	"context"
	"errors"

	"cosmossdk.io/core/transaction"
)

// Mempool defines the required methods of an application's mempool.
type Mempool[T transaction.Tx] interface {
	// Insert attempts to insert a Tx into the app-side mempool returning
	// an error upon failure.
	Insert(context.Context, T) error

	// Select returns an Iterator over the app-side mempool. If txs are specified,
	// then they shall be incorporated into the Iterator. The Iterator must be
	// closed by the caller.
	Select(context.Context, []T) Iterator[T]

	// Remove attempts to remove a transaction from the mempool, returning an error
	// upon failure.
	Remove([]T) error
}

// Iterator defines an app-side mempool iterator interface that is as minimal as
// possible. The order of iteration is determined by the app-side mempool
// implementation.
type Iterator[T transaction.Tx] interface {
	// Next returns the next transaction from the mempool. If there are no more
	// transactions, it returns nil.
	Next() Iterator[T]

	// Tx returns the transaction at the current position of the iterator.
	Tx() T
}

var (
	ErrTxNotFound           = errors.New("tx not found in mempool")
	ErrMempoolTxMaxCapacity = errors.New("pool reached max tx capacity")
)

var _ Mempool[transaction.Tx] = NoOpMempool[transaction.Tx]{}

type NoOpMempool[T transaction.Tx] struct{}

// Insert implements Mempool.
func (NoOpMempool[T]) Insert(context.Context, T) error {
	panic("unimplemented")
}

// Remove implements Mempool.
func (NoOpMempool[T]) Remove([]T) error {
	panic("unimplemented")
}

// Select implements Mempool.
func (NoOpMempool[T]) Select(context.Context, []T) Iterator[T] {
	panic("unimplemented")
}
