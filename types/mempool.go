package types

import (
	"fmt"

	"github.com/google/btree"
)

// MempoolTx we define an app-side mempool transaction interface that is as
// minimal as possible, only requiring applications to define the size of the
// transaction to be used when reaping and getting the transaction itself.
// Interface type casting can be used in the actual app-side mempool implementation.
type MempoolTx interface {
	Tx

	// Size returns the size of the transaction in bytes.
	Size() int
}

// HashableTx defines an interface for a transaction that can be hashed.
// TODO Consider merging with MemPoolTx or using signatures instead.
type HashableTx interface {
	GetHash() [32]byte
}

type Mempool interface {
	// Insert attempts to insert a MempoolTx into the app-side mempool returning
	// an error upon failure.
	Insert(Context, MempoolTx) error

	// Select returns the next set of available transactions from the app-side
	// mempool, up to maxBytes or until the mempool is empty. The application can
	// decide to return transactions from its own mempool, from the incoming
	// txs, or some combination of both.
	Select(ctx Context, txs [][]byte, maxBytes int) ([]MempoolTx, error)

	// CountTx returns the number of transactions currently in the mempool.
	CountTx() int

	// Remove attempts to remove a transaction from the mempool, returning an error
	// upon failure.
	Remove(Context, MempoolTx) error
}

var (
	_ Mempool    = (*btreeMempool)(nil)
	_ btree.Item = (*btreeItem)(nil)
)

var (
	ErrMempoolIsFull = fmt.Errorf("mempool is full")
	ErrNoTxHash      = fmt.Errorf("tx is not hashable")
)

type btreeMempool struct {
	btree      *btree.BTree
	txBytes    int
	maxTxBytes int
	txCount    int
	hashes     map[[32]byte]int64
}

type btreeItem struct {
	// TODO use linked list instead of slice if we opt for a Btree
	txs      []MempoolTx
	priority int64
}

func (bi *btreeItem) Less(than btree.Item) bool {
	return bi.priority < than.(*btreeItem).priority
}

func NewBTreeMempool(maxBytes int) *btreeMempool {
	return &btreeMempool{
		btree:      btree.New(2),
		txBytes:    0,
		maxTxBytes: maxBytes,
		hashes:     make(map[[32]byte]int64),
	}
}

func (btm *btreeMempool) Insert(ctx Context, tx MempoolTx) error {
	hashTx, ok := tx.(HashableTx)
	if !ok {
		return ErrNoTxHash
	}
	txSize := tx.Size()
	priority := ctx.Priority()

	if btm.txBytes+txSize > btm.maxTxBytes {
		return ErrMempoolIsFull
	}

	key := &btreeItem{priority: priority}

	bi := btm.btree.Get(key)
	if bi != nil {
		bi := bi.(*btreeItem)
		bi.txs = append(bi.txs, tx)
	} else {
		bi = &btreeItem{txs: []MempoolTx{tx}, priority: priority}
	}

	btm.btree.ReplaceOrInsert(bi)
	btm.hashes[hashTx.GetHash()] = priority
	btm.txBytes += txSize
	btm.txCount++

	return nil
}

func (btm *btreeMempool) validateSequenceNumber(tx Tx) bool {
	// TODO
	return true
}

func (btm *btreeMempool) Select(_ Context, _ [][]byte, maxBytes int) ([]MempoolTx, error) {
	txBytes := 0
	var selectedTxs []MempoolTx
	btm.btree.Descend(func(i btree.Item) bool {
		txs := i.(*btreeItem).txs
		for _, tx := range txs {
			txSize := tx.Size()
			if txBytes+txSize < maxBytes {
				return false
			}
			if !btm.validateSequenceNumber(tx) {
				continue
			}
			selectedTxs = append(selectedTxs, tx)
			txBytes += txSize
		}

		return true
	})
	return selectedTxs, nil
}

func (btm *btreeMempool) CountTx() int {
	return btm.txCount
}

func (btm *btreeMempool) Remove(_ Context, tx MempoolTx) error {
	hashTx, ok := tx.(HashableTx)
	if !ok {
		return ErrNoTxHash
	}
	hash := hashTx.GetHash()

	priority, txFound := btm.hashes[hash]
	if !txFound {
		return fmt.Errorf("tx %X not found", hash)
	}

	i := btm.btree.Get(&btreeItem{priority: priority})
	if i == nil {
		return fmt.Errorf("tx with priority %v not found", priority)
	}

	item := i.(*btreeItem)
	if len(item.txs) == 1 {
		btm.btree.Delete(i)
	} else {
		found := false
		for j, t := range item.txs {
			if t.(HashableTx).GetHash() == hash {
				item.txs = append(item.txs[:j], item.txs[j+1:]...)
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("tx %X not found at priority %v", hash, priority)
		}
		btm.btree.ReplaceOrInsert(item)
	}

	delete(btm.hashes, hash)
	btm.txBytes -= tx.Size()
	btm.txCount--

	return nil
}
