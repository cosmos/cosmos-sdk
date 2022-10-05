package mempool

import (
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

	senderTxs.Set(tk, tx)
	mp.scores[txKey{nonce: nonce, sender: sender}] = ctx.Priority()
	mp.priorities.Set(tk, tx)

	return nil
}

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

func (mp *defaultMempool) CountTx() int {
	return mp.priorities.Len()
}

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
