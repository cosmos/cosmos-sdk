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

	// Select returns the next set of available transactions from the app-side
	// mempool, up to maxBytes or until the mempool is empty. The application can
	// decide to return transactions from its own mempool, from the incoming
	// txs, or some combination of both.
	Select(txs [][]byte, maxBytes int64) ([]Tx, error)

	// CountTx returns the number of transactions currently in the mempool.
	CountTx() int

	// Remove attempts to remove a transaction from the mempool, returning an error
	// upon failure.
	Remove(Tx) error
}

var ErrTxNotFound = errors.New("tx not found in mempool")

type Factory func() Mempool
