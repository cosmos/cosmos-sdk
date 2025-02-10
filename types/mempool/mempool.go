package mempool

import (
	"context"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Mempool interface {
	// Insert attempts to insert a Tx into the app-side mempool returning
	// an error upon failure.
	Insert(context.Context, sdk.Tx) error

	// Select returns an Iterator over the app-side mempool. If txs are specified,
	// then they shall be incorporated into the Iterator. The Iterator is not thread-safe to use.
	Select(context.Context, [][]byte) Iterator

	// CountTx returns the number of transactions currently in the mempool.
	CountTx() int

	// Remove attempts to remove a transaction from the mempool, returning an error
	// upon failure.
	Remove(sdk.Tx) error
}

// ExtMempool is a extension of Mempool interface introduced in v0.50
// for not be breaking in a patch release.
// In v0.52+, this interface will be merged into Mempool interface.
type ExtMempool interface {
	Mempool

	// SelectBy use callback to iterate over the mempool, it's thread-safe to use.
	SelectBy(context.Context, [][]byte, func(sdk.Tx) bool)
}

// Iterator defines an app-side mempool iterator interface that is as minimal as
// possible. The order of iteration is determined by the app-side mempool
// implementation.
type Iterator interface {
	// Next returns the next transaction from the mempool. If there are no more
	// transactions, it returns nil.
	Next() Iterator

	// Tx returns the transaction at the current position of the iterator.
	Tx() sdk.Tx
}

var (
	ErrTxNotFound           = errors.New("tx not found in mempool")
	ErrMempoolTxMaxCapacity = errors.New("pool reached max tx capacity")
)

// SelectBy is compatible with old interface to avoid breaking api.
// In v0.52+, this function is removed and SelectBy is merged into Mempool interface.
func SelectBy(ctx context.Context, mempool Mempool, txs [][]byte, callback func(sdk.Tx) bool) {
	if ext, ok := mempool.(ExtMempool); ok {
		ext.SelectBy(ctx, txs, callback)
		return
	}

	// fallback to old behavior, without holding the lock while iteration.
	iter := mempool.Select(ctx, txs)
	for iter != nil && callback(iter.Tx()) {
		iter = iter.Next()
	}
}
