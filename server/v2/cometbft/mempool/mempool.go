package mempool

import (
	"context"
	"errors"

	"cosmossdk.io/core/transaction"
)

var (
	ErrTxNotFound           = errors.New("tx not found in mempool")
	ErrMempoolTxMaxCapacity = errors.New("pool reached max tx capacity")
)

// Mempool defines the required methods of an application's mempool.
type Mempool[T transaction.Tx] interface {
	// Insert attempts to insert a Tx into the app-side mempool returning
	// an error upon failure.
	Insert(context.Context, T) error

	// Select returns an Iterator over the app-side mempool. If txs are specified,
	// then they shall be incorporated into the Iterator. The Iterator is not thread-safe to use.
	Select(context.Context, []T) Iterator[T]

	// SelectBy use callback to iterate over the mempool, it's thread-safe to use.
	SelectBy(context.Context, []T, func(T) bool)

	// CountTx returns the number of transactions currently in the mempool.
	CountTx() int

	// Remove attempts to remove a transaction from the mempool, returning an error
	// upon failure.
	Remove(T) error
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
