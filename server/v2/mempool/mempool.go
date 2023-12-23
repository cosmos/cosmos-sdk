package mempool

import (
	"context"

	"cosmossdk.io/server/v2/core/transaction"
)

// Tx defines a mempool tx containing both encoded and decoded version of it.
// This is useful to avoid further encoding and decoding of txs.
type Tx[T transaction.Tx] struct {
	// Decoded is the decoded Tx.
	Decoded T
	// Encoded is the encoded Tx.
	Encoded []byte
}

// Mempool defines the required methods of an application's mempool.
type Mempool[T transaction.Tx] interface {
	// Push pushes the TXs to the mempool.
	Push(ctx context.Context, txs []Tx[T]) error
	// Pull fetches the provided number of txs from the mempool.
	// It is a design detail of the mempool to decide what is the
	// prioritization over the mempool.
	Pull(ctx context.Context, num int) ([]Tx[T], error)
	// Count returns the number of Txs in the mempool.
	Count(ctx context.Context) (int, error)
}
