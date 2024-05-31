package handlers

import (
	"context"

	cmttypes "github.com/cometbft/cometbft/types"

	"cosmossdk.io/core/transaction"
)

// TxSelector defines a helper type that assists in selecting transactions during
// mempool transaction selection in PrepareProposal. It keeps track of the total
// number of bytes and total gas of the selected transactions. It also keeps
// track of the selected transactions themselves.
type TxSelector[T transaction.Tx] interface {
	// SelectedTxs should return a copy of the selected transactions.
	SelectedTxs(ctx context.Context) []T

	// Clear should clear the TxSelector, nulling out all relevant fields.
	Clear()

	// SelectTxForProposal should attempt to select a transaction for inclusion in
	// a proposal based on inclusion criteria defined by the TxSelector. It must
	// return <true> if the caller should halt the transaction selection loop
	// (typically over a mempool) or <false> otherwise.
	SelectTxForProposal(ctx context.Context, maxTxBytes, maxBlockGas uint64, tx T) bool
}

type defaultTxSelector[T transaction.Tx] struct {
	totalTxBytes uint64
	totalTxGas   uint64
	selectedTxs  []T
}

func NewDefaultTxSelector[T transaction.Tx]() TxSelector[T] {
	return &defaultTxSelector[T]{}
}

func (ts *defaultTxSelector[T]) SelectedTxs(_ context.Context) []T {
	txs := make([]T, len(ts.selectedTxs))
	copy(txs, ts.selectedTxs)
	return txs
}

func (ts *defaultTxSelector[T]) Clear() {
	ts.totalTxBytes = 0
	ts.totalTxGas = 0
	ts.selectedTxs = nil
}

func (ts *defaultTxSelector[T]) SelectTxForProposal(_ context.Context, maxTxBytes, maxBlockGas uint64, tx T) bool {
	txSize := uint64(cmttypes.ComputeProtoSizeForTxs([]cmttypes.Tx{tx.Bytes()}))
	txGasLimit, err := tx.GetGasLimit()
	if err != nil {
		return false
	}

	// only add the transaction to the proposal if we have enough capacity
	if (txSize + ts.totalTxBytes) <= maxTxBytes {
		// If there is a max block gas limit, add the tx only if the limit has
		// not been met.
		if maxBlockGas > 0 {
			if (txGasLimit + ts.totalTxGas) <= maxBlockGas {
				ts.totalTxGas += txGasLimit
				ts.totalTxBytes += txSize
				ts.selectedTxs = append(ts.selectedTxs, tx)
			}
		} else {
			ts.totalTxBytes += txSize
			ts.selectedTxs = append(ts.selectedTxs, tx)
		}
	}

	// check if we've reached capacity; if so, we cannot select any more transactions
	return ts.totalTxBytes >= maxTxBytes || (maxBlockGas > 0 && (ts.totalTxGas >= maxBlockGas))
}
