package mempool

import (
	"errors"

	"github.com/cosmos/cosmos-sdk/types"
)

// Tx defines an app-side mempool transaction interface that is as
// minimal as possible, only requiring applications to define the size of the
// transaction to be used when inserting, selecting, and deleting the transaction.
// Interface type casting can be used in the actual app-side mempool implementation.
type Tx interface {
	types.Tx

	// Size returns the size of the transaction in bytes.
	Size() int64
}

type Mempool interface {
	// Insert attempts to insert a Tx into the app-side mempool returning
	// an error upon failure.
	Insert(types.Context, Tx) error

	// Select returns an Iterator over the app-side mempool.  If txs are specified, then they shall be incorporated
	// into the Iterator.  The Iterator must be closed by the caller.
	Select(txs [][]byte) Iterator

	// CountTx returns the number of transactions currently in the mempool.
	CountTx() int

	// Remove attempts to remove a transaction from the mempool, returning an error
	// upon failure.
	Remove(Tx) error
}

// Iterator defines an app-side mempool iterator interface that is as minimal as possible.  The order of iteration
// is determined by the app-side mempool implementation.
type Iterator interface {
	// Next returns the next transaction from the mempool. If there are no more transactions, it returns nil.
	Next() Iterator

	// Tx returns the transaction at the current position of the iterator.
	Tx() Tx
}

var ErrTxNotFound = errors.New("tx not found in mempool")

type Factory func() Mempool
