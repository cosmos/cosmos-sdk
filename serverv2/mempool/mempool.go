package mempool

import (
	"context"

	"cosmossdk.io/core/transaction"
)

// Mempool defines the required methods of an application's mempool.
type Mempool[T any] interface {
	// Insert attempts to insert a Tx into the app-side mempool returning
	// an error upon failure. Insert will validate the transaction using the txValidator
	Insert(ctx context.Context, txs T) error

	// GetTxs returns a list of transactions to add in a block
	// size specifies the size of the block left for transactions
	GetTxs(ctx context.Context, totalSize uint32, txSizeFn TxSizeFn) (ts any, err error)

	// CountTx returns the number of transactions currently in the mempool.
	CountTx() uint32

	// Remove attempts to remove a transaction from the mempool, returning an error
	// upon failure.
	Remove(txs T) error
}

// TxSizeFn defines the function type for calculating the size of a transaction.
type TxSizeFn func(context.Context, transaction.Tx) (uint64, error)
