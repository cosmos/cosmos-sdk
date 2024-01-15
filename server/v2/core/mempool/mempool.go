package mempool

import (
	"context"

	"cosmossdk.io/server/v2/core/transaction"
)

// Mempool defines the required methods of an application's mempool.
type Mempool[T transaction.Tx] interface {
	// Insert attempts to insert a Tx into the app-side mempool returning
	// an error upon failure. Insert will validate the transaction using the txValidator
	Insert(ctx context.Context, txs T) error

	// Get returns a list of transactions to add in a block
	// where num is the number of txs to get. NOTE: size
	// represents the size of a TX in bytes.
	Get(ctx context.Context, size int) ([]T, error)

	// Remove attempts to remove a transaction from the mempool, returning an error
	// upon failure.
	Remove(txs []T) error
}
