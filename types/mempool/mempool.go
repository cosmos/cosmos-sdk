package mempool

import (
	"errors"
	"fmt"

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

func IsEmpty(mempool Mempool) error {
	pmp, ok := mempool.(*priorityMempool)
	if ok {
		if pmp.priorityIndex.Len() != 0 {
			return fmt.Errorf("priorityIndex not empty")
		}

		var countKeys []int64
		for k := range pmp.priorityCounts {
			countKeys = append(countKeys, k)
		}
		for _, k := range countKeys {
			if pmp.priorityCounts[k] != 0 {
				return fmt.Errorf("priorityCounts not zero at %v, got %v", k, pmp.priorityCounts[k])
			}
		}

		var senderKeys []string
		for k := range pmp.senderIndices {
			senderKeys = append(senderKeys, k)
		}
		for _, k := range senderKeys {
			if pmp.senderIndices[k].Len() != 0 {
				return fmt.Errorf("senderIndex not empty for sender %v", k)
			}
		}
		return nil
	}

	smp, ok := mempool.(*senderPriorityMempool)
	if ok {
		if smp.priorityIndex.Len() != 0 {
			return fmt.Errorf("priorityIndex not empty")
		}

		var senderKeys []string
		for k := range smp.senderIndices {
			senderKeys = append(senderKeys, k)
		}
		for _, k := range senderKeys {
			if smp.senderIndices[k].Len() != 0 {
				return fmt.Errorf("senderIndex not empty for sender %v", k)
			}
		}
		return nil
	}

	return fmt.Errorf("unknown mempool type")
}
