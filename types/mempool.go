package types

import (
	"bytes"
	"fmt"

	"github.com/MauriceGit/skiplist"
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
type HashableTx interface {
	MempoolTx
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

// ----------------------------------------------------------------------------
// BTree Implementation
// We use a BTree with degree=2 to approximate a Red-Black Tree.

type btreeMempool struct {
	btree      *btree.BTree
	txBytes    int
	maxTxBytes int
	txCount    int
	scores     map[[32]byte]int64
}

type btreeItem struct {
	tx       HashableTx
	priority int64
}

func (bi *btreeItem) Less(than btree.Item) bool {
	prA := bi.priority
	prB := than.(*btreeItem).priority
	if prA == prB {
		// random, deterministic ordering
		hashA := bi.tx.GetHash()
		hashB := than.(*btreeItem).tx.GetHash()
		return bytes.Compare(hashA[:], hashB[:]) < 0
	}
	return prA < prB
}

func NewBTreeMempool(maxBytes int) *btreeMempool {
	return &btreeMempool{
		btree:      btree.New(2),
		txBytes:    0,
		maxTxBytes: maxBytes,
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

	bi := &btreeItem{priority: priority, tx: hashTx}
	btm.btree.ReplaceOrInsert(bi)

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
		tx := i.(*btreeItem).tx
		if btm.validateSequenceNumber(tx) {
			selectedTxs = append(selectedTxs, tx)

			txBytes += tx.Size()
			if txBytes >= maxBytes {
				return false
			}
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

	priority, txFound := btm.scores[hash]
	if !txFound {
		return fmt.Errorf("tx %X not found", hash)
	}

	res := btm.btree.Delete(&btreeItem{priority: priority, tx: hashTx})
	if res == nil {
		return fmt.Errorf("tx %X not in mempool", hash)
	}

	delete(btm.scores, hash)
	btm.txBytes -= tx.Size()
	btm.txCount--

	return nil
}

// ----------------------------------------------------------------------------
// Skip list implementation

type skipListMempool struct {
	list       *skiplist.SkipList
	txBytes    int
	maxTxBytes int
	scores     map[[32]byte]int64
}

func (item skipListItem) ExtractKey() float64 {
	return float64(item.priority)
}

func (item skipListItem) String() string {
	return fmt.Sprintf("txHash %X", item.tx.(HashableTx).GetHash())
}

type skipListItem struct {
	tx       MempoolTx
	priority int64
}

func (slm *skipListMempool) Insert(ctx Context, tx MempoolTx) error {
	hashTx, ok := tx.(HashableTx)
	if !ok {
		return ErrNoTxHash
	}
	txSize := tx.Size()
	priority := ctx.Priority()

	if slm.txBytes+txSize > slm.maxTxBytes {
		return ErrMempoolIsFull
	}

	item := skipListItem{tx: tx, priority: priority}
	slm.list.Insert(item)
	slm.scores[hashTx.GetHash()] = priority
	slm.txBytes += txSize

	return nil
}

func (slm *skipListMempool) validateSequenceNumber(tx Tx) bool {
	// TODO
	return true
}

func (slm *skipListMempool) Select(_ Context, _ [][]byte, maxBytes int) ([]MempoolTx, error) {
	txBytes := 0
	var selectedTxs []MempoolTx

	n := slm.list.GetLargestNode()
	cnt := slm.list.GetNodeCount()
	for i := 0; i < cnt; i++ {
		tx := n.GetValue().(skipListItem).tx

		if !slm.validateSequenceNumber(tx) {
			continue
		}

		selectedTxs = append(selectedTxs, tx)
		txSize := tx.Size()
		txBytes += txSize
		if txBytes >= maxBytes {
			break
		}

		n = slm.list.Prev(n)
	}

	return selectedTxs, nil
}

func (slm *skipListMempool) CountTx() int {
	return slm.list.GetNodeCount()
}

func (slm *skipListMempool) Remove(_ Context, tx MempoolTx) error {
	hashTx, ok := tx.(HashableTx)
	if !ok {
		return ErrNoTxHash
	}
	hash := hashTx.GetHash()

	priority, txFound := slm.scores[hash]
	if !txFound {
		return fmt.Errorf("tx %X not found", hash)
	}

	item := skipListItem{tx: tx, priority: priority}
	// TODO this is broken.  Key needs hash bytes incorporated to keep it unique
	slm.list.Delete(item)

	delete(slm.scores, hash)
	slm.txBytes -= tx.Size()

	return nil
}
