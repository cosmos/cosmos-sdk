package types

import (
	"bytes"
	"fmt"

	maurice "github.com/MauriceGit/skiplist"
	"github.com/google/btree"
	huandu "github.com/huandu/skiplist"

	"github.com/cosmos/cosmos-sdk/x/auth/signing"
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
// Skip list implementations

// mauriceGit

type mauriceSkipListMempool struct {
	list       *maurice.SkipList
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

func (slm *mauriceSkipListMempool) Insert(ctx Context, tx MempoolTx) error {
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

func (slm *mauriceSkipListMempool) validateSequenceNumber(tx Tx) bool {
	// TODO
	return true
}

func (slm *mauriceSkipListMempool) Select(_ Context, _ [][]byte, maxBytes int) ([]MempoolTx, error) {
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

func (slm *mauriceSkipListMempool) CountTx() int {
	return slm.list.GetNodeCount()
}

func (slm *mauriceSkipListMempool) Remove(_ Context, tx MempoolTx) error {
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

// huandu

type huanduSkipListMempool struct {
	list       *huandu.SkipList
	txBytes    int
	maxTxBytes int
	scores     map[[32]byte]int64
}

type priorityKey struct {
	hash     [32]byte
	priority int64
}

func huanduLess(a, b interface{}) int {
	keyA := a.(priorityKey)
	keyB := b.(priorityKey)
	if keyA.priority == keyB.priority {
		return bytes.Compare(keyA.hash[:], keyB.hash[:])
	} else {
		if keyA.priority < keyB.priority {
			return -1
		} else {
			return 1
		}
	}
}

func NewHuanduSkipListMempool() Mempool {
	list := huandu.New(huandu.LessThanFunc(huanduLess))

	return huanduSkipListMempool{
		list:   list,
		scores: make(map[[32]byte]int64),
	}
}

func (slm huanduSkipListMempool) Insert(ctx Context, tx MempoolTx) error {
	hashTx, ok := tx.(HashableTx)
	if !ok {
		return ErrNoTxHash
	}
	txSize := tx.Size()
	priority := ctx.Priority()

	if slm.txBytes+txSize > slm.maxTxBytes {
		return ErrMempoolIsFull
	}

	hash := hashTx.GetHash()
	key := priorityKey{hash: hash, priority: priority}
	slm.list.Set(key, tx)
	slm.scores[hash] = priority
	slm.txBytes += txSize

	return nil
}

func (slm huanduSkipListMempool) validateSequenceNumber(tx Tx) bool {
	// TODO
	return true
}

func (slm huanduSkipListMempool) Select(_ Context, _ [][]byte, maxBytes int) ([]MempoolTx, error) {
	txBytes := 0
	var selectedTxs []MempoolTx

	n := slm.list.Back()
	for n != nil {
		tx := n.Value.(MempoolTx)

		if !slm.validateSequenceNumber(tx) {
			continue
		}

		selectedTxs = append(selectedTxs, tx)
		txSize := tx.Size()
		txBytes += txSize
		if txBytes >= maxBytes {
			break
		}

		n = n.Prev()
	}

	return selectedTxs, nil
}

func (slm huanduSkipListMempool) CountTx() int {
	return slm.list.Len()
}

func (slm huanduSkipListMempool) Remove(_ Context, tx MempoolTx) error {
	hashTx, ok := tx.(HashableTx)
	if !ok {
		return ErrNoTxHash
	}
	hash := hashTx.GetHash()

	priority, txFound := slm.scores[hash]
	if !txFound {
		return fmt.Errorf("tx %X not found", hash)
	}

	key := priorityKey{hash: hash, priority: priority}
	slm.list.Remove(key)

	delete(slm.scores, hash)
	slm.txBytes -= tx.Size()

	return nil
}

// Statefully ordered mempool

type statefulMempool struct {
	priorities *huandu.SkipList
	senders    map[string]*huandu.SkipList
}

type statefulMempoolTxKey struct {
	nonce    uint64
	priority int64
}

func NewStatefulMempool() Mempool {
	return &statefulMempool{
		priorities: huandu.New(huandu.LessThanFunc(huanduLess)),
		senders:    make(map[string]*huandu.SkipList),
	}
}

func (smp statefulMempool) Insert(ctx Context, tx MempoolTx) error {
	senders := tx.(signing.SigVerifiableTx).GetSigners()
	nonces, err := tx.(signing.SigVerifiableTx).GetSignaturesV2()

	if err != nil {
		return err
	} else if len(senders) != len(nonces) {
		return fmt.Errorf("number of senders (%d) does not match number of nonces (%d)", len(senders), len(nonces))
	}

	// TODO multiple senders
	sender := senders[0].String()
	nonce := nonces[0].Sequence

	senderTxs, ok := smp.senders[sender]
	if !ok {
		senderTxs = huandu.New(huandu.LessThanFunc(func(a, b interface{}) int {
			uint64Compare := huandu.Uint64
			return uint64Compare.Compare(a.(statefulMempoolTxKey).nonce, b.(statefulMempoolTxKey).nonce)
		}))
	}
	senderTxs.Set(nonce, tx)

	key := priorityKey{hash: tx.(HashableTx).GetHash(), priority: ctx.Priority()}
	smp.priorities.Set(key, tx)

	return nil
}

func (smp statefulMempool) Select(_ Context, _ [][]byte, maxBytes int) ([]MempoolTx, error) {
	var selectedTxs []MempoolTx

	// start with the highest priority sender
	curPrio := smp.priorities.Back()
	for curPrio != nil {
		nextPrio := curPrio.Prev()
		// TODO min priority on nil
		// cp := curPrio.Key().(priorityKey).priority
		np := nextPrio.Key().(priorityKey).priority
		// TODO multiple senders
		sender := curPrio.Value.(signing.SigVerifiableTx).GetSigners()[0].String()

		// iterate through the sender's transactions in nonce order
		senderTx := smp.senders[sender].Front()
		for senderTx != nil {
			k := senderTx.Key().(statefulMempoolTxKey)
			// break if we've reached a transaction with a priority lower than the next highest priority in the pool
			if k.priority < np {
				break
			}
			// otherwise, select the transaction and continue iteration
			selectedTxs = append(selectedTxs, senderTx.Value.(MempoolTx))

			senderTx = senderTx.Next()
			// TODO size checking
		}

		curPrio = nextPrio
	}

	return selectedTxs, nil
}

func (smp statefulMempool) CountTx() int {
	//TODO implement me
	panic("implement me")
}

func (smp statefulMempool) Remove(context Context, tx MempoolTx) error {
	//TODO implement me
	// need hash tables retrieving skiplist keys for txs/senders
	panic("implement me")
}
