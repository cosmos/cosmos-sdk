package mempool

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/types"
	"math"

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
}

func txKeyLess(a, b interface{}) int {
	keyA := a.(txKey)
	keyB := b.(txKey)
	res := huandu.Int64.Compare(keyA.priority, keyB.priority)
	if res != 0 {
		return res
	}

	res = huandu.Uint64.Compare(keyA.nonce, keyB.nonce)
	if res != 0 {
		return res
	}

	return huandu.String.Compare(keyA.sender, keyB.sender)
}

func NewDefaultMempool() Mempool {
	return &defaultMempool{
		priorities: huandu.New(huandu.LessThanFunc(txKeyLess)),
		senders:    make(map[string]*huandu.SkipList),
		scores:     make(map[txKey]int64),
	}
}

func (mp defaultMempool) Insert(ctx types.Context, tx Tx) error {
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
	tk := txKey{nonce: nonce, priority: ctx.Priority(), sender: sender}

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
	// TODO for each sender/nonce
	mp.scores[txKey{nonce: nonce, sender: sender}] = ctx.Priority()
	mp.priorities.Set(tk, tx)

	return nil
}

func (mp defaultMempool) Select(_ types.Context, _ [][]byte, maxBytes int) ([]Tx, error) {
	var selectedTxs []Tx
	var txBytes int
	senderCursors := make(map[string]*huandu.Element)

	// start with the highest priority sender
	priorityNode := mp.priorities.Front()
	for priorityNode != nil {
		var nextPriority int64
		nextPriorityNode := priorityNode.Next()
		if nextPriorityNode != nil {
			nextPriority = nextPriorityNode.Key().(txKey).priority
		} else {
			nextPriority = math.MinInt64
		}

		// TODO multiple senders
		// first clear out all txs from *all* senders which have a lower nonce *and* priority greater than or equal to the
		// next priority
		// when processing a tx with multi senders remove it from all other sender queues
		sender := priorityNode.Value.(signing.SigVerifiableTx).GetSigners()[0].String()

		// iterate through the sender's transactions in nonce order
		senderTx, ok := senderCursors[sender]
		if !ok {
			senderTx = mp.senders[sender].Front()
		}

		for senderTx != nil {
			k := senderTx.Key().(txKey)
			// break if we've reached a transaction with a priority lower than the next highest priority in the pool
			if k.priority < nextPriority {
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

func (mp defaultMempool) CountTx() int {
	return mp.priorities.Len()
}

func (mp defaultMempool) Remove(context types.Context, tx Tx) error {
	senders := tx.(signing.SigVerifiableTx).GetSigners()
	nonces, _ := tx.(signing.SigVerifiableTx).GetSignaturesV2()
	// TODO multiple senders
	sender := senders[0].String()
	nonce := nonces[0].Sequence

	// TODO multiple senders
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

	return nil
}
