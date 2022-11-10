package mempool

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Mempool interface {
	// Insert attempts to insert a Tx into the app-side mempool returning
	// an error upon failure.
	Insert(sdk.Context, sdk.Tx) error

	// Select returns an Iterator over the app-side mempool.  If txs are specified, then they shall be incorporated
	// into the Iterator.  The Iterator must be closed by the caller.
	Select(sdk.Context, [][]byte) Iterator

	// CountTx returns the number of transactions currently in the mempool.
	CountTx() int

	// Remove attempts to remove a transaction from the mempool, returning an error
	// upon failure.
	Remove(sdk.Tx) error
}

// Iterator defines an app-side mempool iterator interface that is as minimal as possible.  The order of iteration
// is determined by the app-side mempool implementation.
type Iterator interface {
	// Next returns the next transaction from the mempool. If there are no more transactions, it returns nil.
	Next() Iterator

	// Tx returns the transaction at the current position of the iterator.
	Tx() sdk.Tx
}

var ErrTxNotFound = errors.New("tx not found in mempool")
