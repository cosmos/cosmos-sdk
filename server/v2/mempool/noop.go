package mempool

import (
	"context"

	"cosmossdk.io/core/transaction"
	mempool "cosmossdk.io/server/v2/core/mempool"
)

var _ mempool.Mempool[transaction.Tx] = (*NoOpMempool[transaction.Tx])(nil)

// NoOpMempool defines a no-op mempool. Transactions are completely discarded and
// ignored when BaseApp interacts with the mempool.
//
// Note: When this mempool is used, it assumed that an application will rely
// on CometBFT's transaction ordering defined in `RequestPrepareProposal`, which
// is FIFO-ordered by default.
type NoOpMempool[T transaction.Tx] struct{}

func (NoOpMempool[T]) Insert(context.Context, T) error                 { return nil }
func (NoOpMempool[T]) Select(context.Context, []T) mempool.Iterator[T] { return nil }
func (NoOpMempool[T]) CountTx() int                                    { return 0 }
func (NoOpMempool[T]) Remove([]T) error                                { return nil }
