package mempool

import (
	"fmt"
	"math"

	huandu "github.com/huandu/skiplist"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
)

// Tx we define an app-side mempool transaction interface that is as
// minimal as possible, only requiring applications to define the size of the
// transaction to be used when reaping and getting the transaction itself.
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

type ErrTxNotFound struct {
	error
}

type Factory func() Mempool

var _ Mempool = (*defaultMempool)(nil)

// defaultMempool is the SDK's default mempool implementation which stores txs in a partial ordered set
// by 2 dimensions; priority, and sender-nonce.  Internally it uses one priority ordered skip list and one skip list
// per sender ordered by nonce (sender-nonce).  When there are multiple txs from the same sender, they are not
// always comparable by priority to other sender txs and must be partially ordered by both sender-nonce and priority.
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
}

// txKeyLess is a comparator for txKeys that first compares priority, then nonce, then sender, uniquely identifying
// a transaction.
func txKeyLess(a, b interface{}) int {
	keyA := a.(txKey)
	keyB := b.(txKey)
	res := huandu.Int64.Compare(keyA.priority, keyB.priority)
	if res != 0 {
		return res
	}

	res = huandu.String.Compare(keyA.sender, keyB.sender)
	if res != 0 {
		return res
	}

	return huandu.Uint64.Compare(keyA.nonce, keyB.nonce)
}

// NewDefaultMempool returns the SDK's default mempool implementation which returns txs in a partial order
// by 2 dimensions; priority, and sender-nonce.
func NewDefaultMempool() Mempool {
	return &defaultMempool{
		priorities: huandu.New(huandu.LessThanFunc(txKeyLess)),
		senders:    make(map[string]*huandu.SkipList),
		scores:     make(map[txKey]int64),
	}
}

// Insert attempts to insert a Tx into the app-side mempool in O(log n) time, returning an error if unsuccessful.
// Sender and nonce are derived from the transaction's first signature.
// Transactions are unique by sender and nonce.
// Inserting a duplicate tx is an O(log n) no-op.
// Inserting a duplicate tx with a different priority overwrites the existing tx, changing the total order of
// the mempool.
func (mp *defaultMempool) Insert(ctx types.Context, tx Tx) error {
	sigs, err := tx.(signing.SigVerifiableTx).GetSignaturesV2()
	if err != nil {
		return err
	}
	if len(sigs) == 0 {
		return fmt.Errorf("tx must have at least one signer")
	}

	sig := sigs[0]
	sender := sig.PubKey.Address().String()
	nonce := sig.Sequence
	tk := txKey{nonce: nonce, priority: ctx.Priority(), sender: sender}

	senderTxs, ok := mp.senders[sender]
	// initialize sender mempool if not found
	if !ok {
		senderTxs = huandu.New(huandu.LessThanFunc(func(a, b interface{}) int {
			return huandu.Uint64.Compare(b.(txKey).nonce, a.(txKey).nonce)
		}))
		mp.senders[sender] = senderTxs
	}

	// Since senderTxs is scored by nonce, a changed priority will overwrite the existing txKey.
	senderTxs.Set(tk, tx)

	// Since mp.priorities is scored by priority, then sender, then nonce, a changed priority will create a new key,
	// so we must remove the old key and re-insert it to avoid having the same tx with different priorities indexed
	// twice in the mempool.  This O(log n) remove operation is rare and only happens when a tx's priority changes.
	sk := txKey{nonce: nonce, sender: sender}
	if oldScore, txExists := mp.scores[sk]; txExists {
		mp.priorities.Remove(txKey{nonce: nonce, priority: oldScore, sender: sender})
	}
	mp.scores[sk] = ctx.Priority()
	mp.priorities.Set(tk, tx)

	return nil
}

// Select returns a set of transactions from the mempool, prioritized by priority and sender-nonce in O(n) time.
// The passed in list of transactions are ignored.  This is a readonly operation, the mempool is not modified.
// maxBytes is the maximum number of bytes of transactions to return.
func (mp *defaultMempool) Select(_ [][]byte, maxBytes int64) ([]Tx, error) {
	var selectedTxs []Tx
	var txBytes int64
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

// CountTx returns the number of transactions in the mempool.
func (mp *defaultMempool) CountTx() int {
	return mp.priorities.Len()
}

// Remove removes a transaction from the mempool in O(log n) time, returning an error if unsuccessful.
func (mp *defaultMempool) Remove(tx Tx) error {
	sigs, err := tx.(signing.SigVerifiableTx).GetSignaturesV2()
	if err != nil {
		return err
	}
	if len(sigs) == 0 {
		return fmt.Errorf("attempted to remove a tx with no signatures")
	}
	sig := sigs[0]
	sender := sig.PubKey.Address().String()
	nonce := sig.Sequence

	sk := txKey{nonce: nonce, sender: sender}
	priority, ok := mp.scores[sk]
	if !ok {
		return ErrTxNotFound{fmt.Errorf("tx %v not found in mempool", tx)}
	}
	tk := txKey{nonce: nonce, priority: priority, sender: sender}

	senderTxs, ok := mp.senders[sender]
	if !ok {
		return fmt.Errorf("sender %s not found", sender)
	}

	mp.priorities.Remove(tk)
	senderTxs.Remove(tk)
	delete(mp.scores, sk)

	return nil
}

func DebugPrintKeys(mempool Mempool) {
	mp := mempool.(*defaultMempool)
	n := mp.priorities.Front()
	for n != nil {
		k := n.Key().(txKey)
		fmt.Printf("%s, %d, %d\n", k.sender, k.priority, k.nonce)
		n = n.Next()
	}
}

func DebugIterations(mempool Mempool) int {
	mp := mempool.(*defaultMempool)
	return mp.iterations
}
