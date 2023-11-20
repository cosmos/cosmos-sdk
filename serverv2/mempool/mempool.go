package mempool

import (
	"context"
)

// Mempool defines the required methods of an application's mempool.
type Mempool interface {
	// Insert attempts to insert a Tx into the app-side mempool returning
	// an error upon failure. Insert will validate the transaction using the txValidator
	Insert(ctx context.Context, txs any) error

	// GetTxs returns a list of transactions to add in a block
	// size specifies the size of the block left for transactions
	GetTxs(ctx context.Context, size uint32) (ts any, err error)

	// CountTx returns the number of transactions currently in the mempool.
	CountTx() uint32

	// Remove attempts to remove a transaction from the mempool, returning an error
	// upon failure.
	Remove(txs any) error
}
