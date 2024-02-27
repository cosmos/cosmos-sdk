package mempool

import (
	"context"

	"cosmossdk.io/core/transaction"
	"cosmossdk.io/server/v2/core/mempool"
)

// Mempool defines the required methods of an application's mempool.
type Mempool[T transaction.Tx] interface {
	// Insert attempts to insert a Tx into the app-side mempool returning
	// an error upon failure.
	Insert(context.Context, T) error

	// Select returns an Iterator over the app-side mempool. If txs are specified,
	// then they shall be incorporated into the Iterator. The Iterator must be
	// closed by the caller.
	Select(context.Context, []T) mempool.Iterator[T]

	// Remove attempts to remove a transaction from the mempool, returning an error
	// upon failure.
	Remove([]T) error
}
