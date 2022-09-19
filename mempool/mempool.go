package mempool

import (
	"bytes"
	"fmt"
	"github.com/cosmos/cosmos-sdk/types"
	"math"

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
	types.Tx

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
	Insert(types.Context, MempoolTx) error

	// Select returns the next set of available transactions from the app-side
	// mempool, up to maxBytes or until the mempool is empty. The application can
	// decide to return transactions from its own mempool, from the incoming
	// txs, or some combination of both.
	Select(ctx types.Context, txs [][]byte, maxBytes int) ([]MempoolTx, error)

	// CountTx returns the number of transactions currently in the mempool.
	CountTx() int

	// Remove attempts to remove a transaction from the mempool, returning an error
	// upon failure.
	Remove(types.Context, MempoolTx) error
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

func (btm *btreeMempool) Insert(ctx types.Context, tx MempoolTx) error {
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

func (btm *btreeMempool) validateSequenceNumber(tx types.Tx) bool {
	// TODO
	return true
}

func (btm *btreeMempool) Select(_ types.Context, _ [][]byte, maxBytes int) ([]MempoolTx, error) {
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

func (btm *btreeMempool) Remove(_ types.Context, tx MempoolTx) error {
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

func (slm *mauriceSkipListMempool) Insert(ctx types.Context, tx MempoolTx) error {
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

func (slm *mauriceSkipListMempool) validateSequenceNumber(tx types.Tx) bool {
	// TODO
	return true
}

func (slm *mauriceSkipListMempool) Select(_ types.Context, _ [][]byte, maxBytes int) ([]MempoolTx, error) {
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

func (slm *mauriceSkipListMempool) Remove(_ types.Context, tx MempoolTx) error {
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

func (slm huanduSkipListMempool) Insert(ctx types.Context, tx MempoolTx) error {
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

func (slm huanduSkipListMempool) validateSequenceNumber(tx types.Tx) bool {
	// TODO
	return true
}

func (slm huanduSkipListMempool) Select(_ types.Context, _ [][]byte, maxBytes int) ([]MempoolTx, error) {
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

func (slm huanduSkipListMempool) Remove(_ types.Context, tx MempoolTx) error {
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
	scores     map[[32]byte]int64
}

type statefulMempoolTxKey struct {
	nonce    uint64
	priority int64
}

type statefullPriorityKey struct {
	hash     [32]byte
	priority int64
	nonce    uint64
}

func statefullHuanduLess(a, b interface{}) int {
	keyA := a.(statefullPriorityKey)
	keyB := b.(statefullPriorityKey)
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

func NewStatefulMempool() Mempool {
	return &statefulMempool{
		priorities: huandu.New(huandu.LessThanFunc(statefullHuanduLess)),
		senders:    make(map[string]*huandu.SkipList),
		scores:     make(map[[32]byte]int64),
	}
}

func (smp statefulMempool) Insert(ctx types.Context, tx MempoolTx) error {
	senders := tx.(signing.SigVerifiableTx).GetSigners()
	nonces, err := tx.(signing.SigVerifiableTx).GetSignaturesV2()
	hashTx, ok := tx.(HashableTx)
	if !ok {
		return ErrNoTxHash
	}

	if err != nil {
		return err
	} else if len(senders) != len(nonces) {
		return fmt.Errorf("number of senders (%d) does not match number of nonces (%d)", len(senders), len(nonces))
	}

	// TODO multiple senders
	sender := senders[0].String()
	nonce := nonces[0].Sequence

	senderTxs, ok := smp.senders[sender]
	// initialize sender mempool if not found
	if !ok {
		senderTxs = huandu.New(huandu.LessThanFunc(func(a, b interface{}) int {
			uint64Compare := huandu.Uint64
			return uint64Compare.Compare(a.(statefulMempoolTxKey).nonce, b.(statefulMempoolTxKey).nonce)
		}))
	}
	senderKey := statefulMempoolTxKey{nonce: nonce, priority: ctx.Priority()}
	// if a tx with the same nonce exists, replace it and delete from the priority list
	nonceTx := senderTxs.Get(senderKey)
	if nonceTx != nil {
		h := nonceTx.Value.(HashableTx).GetHash()
		smp.priorities.Remove(priorityKey{hash: h, priority: smp.scores[h]})
		if err != nil {
			return err
		}
	}

	senderTxs.Set(senderKey, tx)
	smp.senders[sender] = senderTxs
	key := statefullPriorityKey{hash: hashTx.GetHash(), priority: ctx.Priority(), nonce: nonce}

	smp.priorities.Set(key, tx)
	smp.scores[key.hash] = key.priority

	return nil
}

func (smp statefulMempool) Select(_ types.Context, _ [][]byte, maxBytes int) ([]MempoolTx, error) {
	var selectedTxs []MempoolTx
	var txBytes int

	// start with the highest priority sender
	priorityNode := smp.priorities.Front()
	for priorityNode != nil {
		var nextPriority int64
		nextPriorityNode := priorityNode.Next()
		if nextPriorityNode != nil {
			nextPriority = nextPriorityNode.Key().(statefullPriorityKey).priority
		} else {
			nextPriority = math.MinInt64
		}

		// TODO multiple senders
		sender := priorityNode.Value.(signing.SigVerifiableTx).GetSigners()[0].String()

		priorityNodeKey := priorityNode.Key().(statefullPriorityKey)

		// iterate through the sender's transactions in nonce order
		senderTxPool, ok := smp.senders[sender]
		if !ok {
			return []MempoolTx{}, fmt.Errorf("sender does not exist")
		}
		senderTx := senderTxPool.Back()
		for senderTx != nil {
			key := senderTx.Key().(statefulMempoolTxKey)
			if key.nonce != priorityNodeKey.nonce {
				break
			}
			// break if we've reached a transaction with a priority lower than the next highest priority in the pool
			if key.priority < nextPriority {
				break
			}

			mempoolTx, _ := senderTx.Value.(MempoolTx)
			// otherwise, select the transaction and continue iteration
			selectedTxs = append(selectedTxs, mempoolTx)
			if txBytes += mempoolTx.Size(); txBytes >= maxBytes {
				return selectedTxs, nil
			}
			senderTx = senderTx.Prev()
			senderTxPool.Remove(key)
		}

		priorityNode = nextPriorityNode
	}

	return selectedTxs, nil
}

func (smp statefulMempool) CountTx() int {
	return smp.priorities.Len()
}

func (smp statefulMempool) Remove(context types.Context, tx MempoolTx) error {
	senders := tx.(signing.SigVerifiableTx).GetSigners()
	nonces, _ := tx.(signing.SigVerifiableTx).GetSignaturesV2()

	hashTx, ok := tx.(HashableTx)
	if !ok {
		return ErrNoTxHash
	}
	txHash := hashTx.GetHash()

	priority, ok := smp.scores[txHash]
	if !ok {
		return fmt.Errorf("tx %X not found", txHash)
	}

	// TODO multiple senders
	sender := senders[0].String()
	nonce := nonces[0].Sequence

	senderTxs, ok := smp.senders[sender]
	if !ok {
		return fmt.Errorf("sender %s not found", sender)
	}

	smp.priorities.Remove(priorityKey{hash: txHash, priority: priority})
	senderTxs.Remove(nonce)
	delete(smp.scores, txHash)

	return nil
}

////////////////////

type MemPoolI struct {
	accountsHeads *huandu.SkipList
	senders       map[string]*AccountMemPool
}

type AccountMemPool struct {
	transactions *huandu.SkipList
	currentKey   statefullPriorityKey
	currentItem  *huandu.Element
}

func priorityHuanduLess(a, b interface{}) int {
	keyA := a.(statefullPriorityKey)
	keyB := b.(statefullPriorityKey)
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

func nonceHuanduLess(a, b interface{}) int {
	keyA := a.(statefullPriorityKey)
	keyB := b.(statefullPriorityKey)
	uint64Compare := huandu.Uint64
	return uint64Compare.Compare(keyA.nonce, keyB.nonce)
}

func priorityHuanduLess2(a, b interface{}) int {
	keyA := a.(accountsHeadsKey)
	keyB := b.(accountsHeadsKey)
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

type accountsHeadsKey struct {
	sender   string
	priority int64
	hash     [32]byte
}

func NewMemPoolI() MemPoolI {
	return MemPoolI{
		accountsHeads: huandu.New(huandu.LessThanFunc(priorityHuanduLess2)),
		senders:       make(map[string]*AccountMemPool),
	}
}

func (amp *MemPoolI) Insert(ctx types.Context, tx MempoolTx) error {
	senders := tx.(signing.SigVerifiableTx).GetSigners()
	nonces, err := tx.(signing.SigVerifiableTx).GetSignaturesV2()
	hashTx, ok := tx.(HashableTx)
	if !ok {
		return ErrNoTxHash
	}

	if err != nil {
		return err
	} else if len(senders) != len(nonces) {
		return fmt.Errorf("number of senders (%d) does not match number of nonces (%d)", len(senders), len(nonces))
	}
	sender := senders[0].String()
	nonce := nonces[0].Sequence

	accountMeempool, ok := amp.senders[sender]
	if !ok {
		accountMeempool = &AccountMemPool{
			transactions: huandu.New(huandu.LessThanFunc(nonceHuanduLess)),
		}
	}
	key := statefullPriorityKey{hash: hashTx.GetHash(), nonce: nonce, priority: ctx.Priority()}

	accountMeempool.transactions.Set(key, tx)
	accKey := accountsHeadsKey{sender: sender, priority: ctx.Priority(), hash: hashTx.GetHash()}

	newTopTx := accountMeempool.transactions.Front()
	accountMeempool.currentItem = newTopTx
	amp.accountsHeads.Remove(accountMeempool.currentKey)
	accountMeempool.currentKey =
		amp.accountsHeads.Set(accKey, accountMeempool)
	amp.senders[sender] = accountMeempool
	return nil

}

func (amp *MemPoolI) Select(_ types.Context, _ [][]byte, maxBytes int) ([]MempoolTx, error) {
	var selectedTxs []MempoolTx
	var txBytes int

	currentAccount := amp.accountsHeads.Front()
	for currentAccount != nil {
		accountMemPool := currentAccount.Value.(*AccountMemPool)
		//currentTx := accountMemPool.transactions.Front()
		currentTx := accountMemPool.currentItem
		tx := currentTx.Value.(MempoolTx)
		selectedTxs = append(selectedTxs, tx)
		if txBytes += tx.Size(); txBytes >= maxBytes {
			return selectedTxs, nil
		}
		newCurrentTx := currentTx.Next()
		accountMemPool.currentItem = newCurrentTx
		accountMemPool.currentKey = newCurrentTx.Key().(statefullPriorityKey)
		//accountMemPool.transactions.Remove(currentTx.Key())
		amp.accountsHeads.Remove(currentAccount.Key())
		amp.accountsHeads.Set(accountMemPool.currentKey, accountMemPool)
		currentAccount = amp.accountsHeads.Front()
	}
	return selectedTxs, nil
}
