package mempool

import (
	"bytes"
	"fmt"
	"math"

	"github.com/cosmos/cosmos-sdk/types"

	huandu "github.com/huandu/skiplist"

	"github.com/cosmos/cosmos-sdk/x/auth/signing"
)

// Tx we define an app-side mempool transaction interface that is as
// minimal as possible, only requiring applications to define the size of the
// transaction to be used when reaping and getting the transaction itself.
// Interface type casting can be used in the actual app-side mempool implementation.
type Tx interface {
	types.Tx

	// Size returns the size of the transaction in bytes.
	Size() int
}

// HashableTx defines an interface for a transaction that can be hashed.
type HashableTx interface {
	Tx
	GetHash() [32]byte
}

type Mempool interface {
	// Insert attempts to insert a Tx into the app-side mempool returning
	// an error upon failure.
	Insert(types.Context, Tx) error

	// Select returns the next set of available transactions from the app-side
	// mempool, up to maxBytes or until the mempool is empty. The application can
	// decide to return transactions from its own mempool, from the incoming
	// txs, or some combination of both.
	Select(ctx types.Context, txs [][]byte, maxBytes int) ([]Tx, error)

	// CountTx returns the number of transactions currently in the mempool.
	CountTx() int

	// Remove attempts to remove a transaction from the mempool, returning an error
	// upon failure.
	Remove(types.Context, Tx) error
}

var (
	_ Mempool = (*defaultMempool)(nil)
)

var (
	ErrMempoolIsFull = fmt.Errorf("mempool is full")
	ErrNoTxHash      = fmt.Errorf("tx is not hashable")
)

type defaultMempool struct {
	priorities *huandu.SkipList
	senders    map[string]*huandu.SkipList
	scores     map[txKey]int64
	iterations int
}

type txKey struct {
	nonce    uint64
	priority int64
	sender   string
	hash     [32]byte
}

func txKeyLess(a, b interface{}) int {
	keyA := a.(txKey)
	keyB := b.(txKey)
	res := huandu.Int64.Compare(keyA.priority, keyB.priority)
	if res != 0 {
		return res
	}

	res = bytes.Compare(keyB.hash[:], keyA.hash[:])
	//res = huandu.Bytes.Compare(keyA.hash[:], keyB.hash[:])
	if res != 0 {
		return res
	}

	res = huandu.String.Compare(keyA.sender, keyB.sender)
	if res != 0 {
		return res
	}

	return huandu.Uint64.Compare(keyA.nonce, keyB.nonce)
}

func NewDefaultMempool() Mempool {
	return &defaultMempool{
		priorities: huandu.New(huandu.LessThanFunc(txKeyLess)),
		senders:    make(map[string]*huandu.SkipList),
		scores:     make(map[txKey]int64),
	}
}

func (mp *defaultMempool) Insert(ctx types.Context, tx Tx) error {
	sigs, err := tx.(signing.SigVerifiableTx).GetSignaturesV2()
	if err != nil {
		return err
	}

	// TODO better selection criteria here
	sig := sigs[0]
	sender := sig.PubKey.Address().String()
	nonce := sig.Sequence

	hashableTx, ok := tx.(HashableTx)
	if !ok {
		return ErrNoTxHash
	}

	tk := txKey{nonce: nonce, priority: ctx.Priority(), sender: sender, hash: hashableTx.GetHash()}

	senderTxs, ok := mp.senders[sender]
	// initialize sender mempool if not found
	if !ok {
		senderTxs = huandu.New(huandu.LessThanFunc(func(a, b interface{}) int {
			return huandu.Uint64.Compare(b.(txKey).nonce, a.(txKey).nonce)
		}))
		mp.senders[sender] = senderTxs
	}

	// if a tx with the same nonce exists, replace it and delete from the priority list
	senderTxs.Set(tk, tx)
	mp.scores[txKey{nonce: nonce, sender: sender}] = ctx.Priority()
	mp.priorities.Set(tk, tx)

	return nil
}

func (mp *defaultMempool) Select(_ types.Context, _ [][]byte, maxBytes int) ([]Tx, error) {
	var selectedTxs []Tx
	var txBytes int
	senderCursors := make(map[string]*huandu.Element)

	// start with the highest priority sender
	priorityNode := mp.priorities.Front()
	for priorityNode != nil {
		priorityKey := priorityNode.Key().(txKey)
		nextHighestPriority, nextPriorityNode := nextPriority(priorityNode)
		sender := priorityKey.sender
		senderTx := mp.fetchSenderCursor(senderCursors, sender)

		// iterate through the sender's transactions in nonce order
		for senderTx != nil {
			// time complexity tracking
			mp.iterations++
			k := senderTx.Key().(txKey)

			// break if we've reached a transaction with a priority lower than the next highest priority in the pool
			if k.priority < nextHighestPriority {
				break
			}

			mempoolTx, _ := senderTx.Value.(Tx)
			// otherwise, select the transaction and continue iteration
			selectedTxs = append(selectedTxs, mempoolTx)
			if txBytes += mempoolTx.Size(); txBytes >= maxBytes {
				return selectedTxs, nil
			}

			senderTx = senderTx.Next()
			senderCursors[sender] = senderTx
		}

		priorityNode = nextPriorityNode
	}

	return selectedTxs, nil
}

func (mp *defaultMempool) fetchSenderCursor(senderCursors map[string]*huandu.Element, sender string) *huandu.Element {
	senderTx, ok := senderCursors[sender]
	if !ok {
		senderTx = mp.senders[sender].Front()
	}
	return senderTx
}

func nextPriority(priorityNode *huandu.Element) (int64, *huandu.Element) {
	var np int64
	nextPriorityNode := priorityNode.Next()
	if nextPriorityNode != nil {
		np = nextPriorityNode.Key().(txKey).priority
	} else {
		np = math.MinInt64
	}
	return np, nextPriorityNode
}

func (mp *defaultMempool) CountTx() int {
	return mp.priorities.Len()
}

func (mp *defaultMempool) Remove(context types.Context, tx Tx) error {
	senders := tx.(signing.SigVerifiableTx).GetSigners()
	nonces, _ := tx.(signing.SigVerifiableTx).GetSignaturesV2()

	for i, senderAddr := range senders {
		sender := senderAddr.String()
		nonce := nonces[i].Sequence

		tk := txKey{sender: sender, nonce: nonce}

		priority, ok := mp.scores[tk]
		if !ok {
			return fmt.Errorf("tx %v not found", tk)
		}

		senderTxs, ok := mp.senders[sender]
		if !ok {
			return fmt.Errorf("sender %s not found", sender)
		}

		mp.priorities.Remove(txKey{priority: priority, sender: sender, nonce: nonce})
		senderTxs.Remove(nonce)
		delete(mp.scores, tk)
	}
	return nil
}

func DebugPrintKeys(mempool Mempool) {
	mp := mempool.(*defaultMempool)
	n := mp.priorities.Front()
	for n != nil {
		k := n.Key().(txKey)
		fmt.Printf("%s, %d, %d; %d\n", k.sender, k.priority, k.nonce, k.hash[0])
		n = n.Next()
	}
}

func Iterations(mempool Mempool) int {
	switch v := mempool.(type) {
	case *defaultMempool:
		return v.iterations
	case *graph:
		return v.iterations
	}
	panic("unknown mempool type")
}

// The complexity is O(log(N)). Implementation
type statefullPriorityKey struct {
	hash     [32]byte
	priority int64
	nonce    uint64
}

type accountsHeadsKey struct {
	sender   string
	priority int64
	hash     [32]byte
}

type AccountMemPool struct {
	transactions *huandu.SkipList
	currentKey   accountsHeadsKey
	currentItem  *huandu.Element
	sender       string
}

// Push cannot be executed in the middle of a select
func (amp *AccountMemPool) Push(ctx types.Context, key statefullPriorityKey, tx Tx) {
	amp.transactions.Set(key, tx)
	amp.currentItem = amp.transactions.Back()
	newKey := amp.currentItem.Key().(statefullPriorityKey)
	amp.currentKey = accountsHeadsKey{hash: newKey.hash, sender: amp.sender, priority: newKey.priority}
}

func (amp *AccountMemPool) Pop() *Tx {
	if amp.currentItem == nil {
		return nil
	}
	itemToPop := amp.currentItem
	amp.currentItem = itemToPop.Prev()
	if amp.currentItem != nil {
		newKey := amp.currentItem.Key().(statefullPriorityKey)
		amp.currentKey = accountsHeadsKey{hash: newKey.hash, sender: amp.sender, priority: newKey.priority}
	} else {
		amp.currentKey = accountsHeadsKey{}
	}
	tx := itemToPop.Value.(Tx)
	return &tx
}

type MemPoolI struct {
	accountsHeads *huandu.SkipList
	senders       map[string]*AccountMemPool
}

func NewMemPoolI() MemPoolI {
	return MemPoolI{
		accountsHeads: huandu.New(huandu.LessThanFunc(priorityHuanduLess)),
		senders:       make(map[string]*AccountMemPool),
	}
}

func (amp *MemPoolI) Insert(ctx types.Context, tx Tx) error {
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
			sender:       sender,
		}
	}
	key := statefullPriorityKey{hash: hashTx.GetHash(), nonce: nonce, priority: ctx.Priority()}

	prevKey := accountMeempool.currentKey
	accountMeempool.Push(ctx, key, tx)

	amp.accountsHeads.Remove(prevKey)
	amp.accountsHeads.Set(accountMeempool.currentKey, accountMeempool)
	amp.senders[sender] = accountMeempool
	return nil

}

func (amp *MemPoolI) Select(_ types.Context, _ [][]byte, maxBytes int) ([]Tx, error) {
	var selectedTxs []Tx
	var txBytes int

	currentAccount := amp.accountsHeads.Front()
	for currentAccount != nil {
		accountMemPool := currentAccount.Value.(*AccountMemPool)
		//currentTx := accountMemPool.transactions.Front()
		prevKey := accountMemPool.currentKey
		tx := accountMemPool.Pop()
		if tx == nil {
			return selectedTxs, nil
		}
		mempoolTx := *tx
		selectedTxs = append(selectedTxs, mempoolTx)
		if txBytes += mempoolTx.Size(); txBytes >= maxBytes {
			return selectedTxs, nil
		}

		amp.accountsHeads.Remove(prevKey)
		amp.accountsHeads.Set(accountMemPool.currentKey, accountMemPool)
		currentAccount = amp.accountsHeads.Front()
	}
	return selectedTxs, nil
}

func priorityHuanduLess(a, b interface{}) int {
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

func nonceHuanduLess(a, b interface{}) int {
	keyA := a.(statefullPriorityKey)
	keyB := b.(statefullPriorityKey)
	uint64Compare := huandu.Uint64
	return uint64Compare.Compare(keyA.nonce, keyB.nonce)
}
