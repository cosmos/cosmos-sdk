package mock

import (
	"context"

	"cosmossdk.io/core/transaction"
	"cosmossdk.io/server/v2/cometbft/mempool"
)

var _ mempool.Mempool[transaction.Tx] = (*MockMempool[transaction.Tx])(nil)

// MockMempool defines a no-op mempool. Transactions are completely discarded and
// ignored when BaseApp interacts with the mempool.
//
// Note: When this mempool is used, it assumed that an application will rely
// on CometBFT's transaction ordering defined in `RequestPrepareProposal`, which
// is FIFO-ordered by default.
type MockMempool[T transaction.Tx] struct{}

func (MockMempool[T]) Insert(context.Context, T) error         { return nil }
func (MockMempool[T]) Select(context.Context, []T) mempool.Iterator[T] { return nil }
func (MockMempool[T]) CountTx() int                            { return 0 }
func (MockMempool[T]) Remove([]T) error                        { return nil }
