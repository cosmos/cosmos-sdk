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
// TODO Consider merging with MemPoolTx.
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

var _ Mempool = (*btreeMempool)(nil)
var _ btree.Item = (*btreeItem)(nil)

type btreeMempool struct {
	btree      *btree.BTree
	txBytes    int
	maxTxBytes int
	txCount    int
	hashes     map[[32]byte]int64
}

type btreeItem struct {
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
	}
}

var ErrMempoolIsFull = fmt.Errorf("mempool is full")
var ErrNoTxHash = fmt.Errorf("tx is not hashable")

func (btm *btreeMempool) Insert(ctx Context, tx MempoolTx) error {
	hashTx, ok := tx.(HashableTx)
	if !ok {
		return ErrNoTxHash
	}
	txSize := tx.Size()
	if btm.txBytes+txSize > btm.maxTxBytes {
		return ErrMempoolIsFull
	}

	var bi *btreeItem
	key := &btreeItem{priority: ctx.Priority()}
	if btm.btree.Has(key) {
		bi = btm.btree.Get(key).(*btreeItem)
		bi.txs = append(bi.txs, tx)
	} else {
		bi = &btreeItem{txs: []MempoolTx{tx}, priority: ctx.Priority()}
	}

	btm.btree.ReplaceOrInsert(bi)
	btm.hashes[hashTx.GetHash()] = ctx.Priority()
	btm.txBytes += txSize
	btm.txCount++

	return nil
}

func (btm *btreeMempool) Select(ctx Context, _ [][]byte, maxBytes int) ([]MempoolTx, error) {
	// TODO sequence no. validation

	txBytes := 0
	var selectedTxs []MempoolTx
	btm.btree.Descend(func(i btree.Item) bool {
		txs := i.(*btreeItem).txs
		for _, tx := range txs {
			txBytes += tx.Size()
			if txBytes < maxBytes {
				return false
			}
			selectedTxs = append(selectedTxs, tx)
		}

		return true
	})
	return selectedTxs, nil
}

func (btm *btreeMempool) CountTx() int {
	return btm.txCount
}

func (btm *btreeMempool) Remove(context Context, tx MempoolTx) error {
	hashTx, ok := tx.(HashableTx)
	if !ok {
		return ErrNoTxHash
	}
	hash := hashTx.GetHash()

	priority, txFound := btm.hashes[hash]
	if !txFound {
		return fmt.Errorf("tx %X not found", hash)
	}

	// TODO handle tx arrays
	i := btm.btree.Delete(&btreeItem{priority: priority})
	if i == nil {
		return fmt.Errorf("tx with priority %v not found", priority)
	}

	delete(btm.hashes, hash)
	return nil
}
