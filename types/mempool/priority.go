package mempool

import (
	"fmt"
	"math"

	huandu "github.com/huandu/skiplist"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
)

var _ Mempool = (*priorityMempool)(nil)

// priorityMempool is the SDK's default mempool implementation which stores txs in a partially ordered set
// by 2 dimensions: priority, and sender-nonce (sequence number).  Internally it uses one priority ordered skip list
// and one skip per sender ordered by nonce (sender-nonce).  When there are multiple txs from the same sender,
// they are not always comparable by priority to other sender txs and must be partially ordered by both sender-nonce
// and priority.
type priorityMempool struct {
	priorityIndex  *huandu.SkipList
	priorityCounts map[int64]int
	senderIndices  map[string]*huandu.SkipList
	senderCursors  map[string]*huandu.Element
	scores         map[txMeta]txMeta
	onRead         func(tx Tx)
}

// txMeta stores transaction metadata used in indices
type txMeta struct {
	nonce         uint64
	priority      int64
	sender        string
	weight        int64
	senderElement *huandu.Element
}

// txMetaLess is a comparator for txKeys that first compares priority, then sender, then nonce, uniquely identifying
// a transaction.
func txMetaLess(a, b any) int {
	keyA := a.(txMeta)
	keyB := b.(txMeta)
	res := huandu.Int64.Compare(keyA.priority, keyB.priority)
	if res != 0 {
		return res
	}

	res = huandu.Int64.Compare(keyA.weight, keyB.weight)
	if res != 0 {
		return res
	}

	res = huandu.String.Compare(keyA.sender, keyB.sender)
	if res != 0 {
		return res
	}

	return huandu.Uint64.Compare(keyA.nonce, keyB.nonce)
}

type PriorityMempoolOption func(*priorityMempool)

// WithOnRead sets a callback to be called when a tx is read from the mempool.
func WithOnRead(onRead func(tx Tx)) PriorityMempoolOption {
	return func(mp *priorityMempool) {
		mp.onRead = onRead
	}
}

// DefaultPriorityMempool returns a priorityMempool with no options.
func DefaultPriorityMempool() Mempool {
	return NewPriorityMempool()
}

// NewPriorityMempool returns the SDK's default mempool implementation which returns txs in a partial order
// by 2 dimensions; priority, and sender-nonce.
func NewPriorityMempool(opts ...PriorityMempoolOption) Mempool {
	mp := &priorityMempool{
		priorityIndex:  huandu.New(huandu.LessThanFunc(txMetaLess)),
		priorityCounts: make(map[int64]int),
		senderIndices:  make(map[string]*huandu.SkipList),
		senderCursors:  make(map[string]*huandu.Element),
		scores:         make(map[txMeta]txMeta),
	}
	for _, opt := range opts {
		opt(mp)
	}
	return mp
}

// Insert attempts to insert a Tx into the app-side mempool in O(log n) time, returning an error if unsuccessful.
// Sender and nonce are derived from the transaction's first signature.
// Transactions are unique by sender and nonce.
// Inserting a duplicate tx is an O(log n) no-op.
// Inserting a duplicate tx with a different priority overwrites the existing tx, changing the total order of
// the mempool.
func (mp *priorityMempool) Insert(ctx sdk.Context, tx Tx) error {
	sigs, err := tx.(signing.SigVerifiableTx).GetSignaturesV2()
	if err != nil {
		return err
	}
	if len(sigs) == 0 {
		return fmt.Errorf("tx must have at least one signer")
	}

	priority := ctx.Priority()
	sig := sigs[0]
	sender := sig.PubKey.Address().String()
	nonce := sig.Sequence
	tk := txMeta{nonce: nonce, priority: priority, sender: sender}
	weight := int64(0)

	senderIndex, ok := mp.senderIndices[sender]
	if !ok {
		senderIndex = huandu.New(huandu.LessThanFunc(func(a, b any) int {
			return huandu.Uint64.Compare(b.(txMeta).nonce, a.(txMeta).nonce)
		}))
		// initialize sender index if not found
		mp.senderIndices[sender] = senderIndex
	}

	mp.priorityCounts[priority] = mp.priorityCounts[priority] + 1
	// Since senderIndex is scored by nonce, a changed priority will overwrite the existing txKey.
	senderElement := senderIndex.Set(tk, tx)
	if mp.priorityCounts[priority] > 1 {
		// needs a weight
		weight = senderWeight(senderElement)
	}
	if senderElement.Prev() != nil && mp.priorityCounts[senderElement.Prev().Key().(txMeta).priority] > 1 {
		// previous element needs its weight updated
		// delete/insert
	}

	tk.weight = weight

	// Since mp.priorityIndex is scored by priority, then sender, then nonce, a changed priority will create a new key,
	// so we must remove the old key and re-insert it to avoid having the same tx with different priorityIndex indexed
	// twice in the mempool.  This O(log n) remove operation is rare and only happens when a tx's priority changes.
	sk := txMeta{nonce: nonce, sender: sender}
	if oldScore, txExists := mp.scores[sk]; txExists {
		mp.priorityIndex.Remove(txMeta{
			nonce:    nonce,
			priority: oldScore.priority,
			sender:   sender,
			weight:   oldScore.weight,
		})
	}
	mp.scores[sk] = txMeta{priority: priority, weight: weight}
	tk.senderElement = senderElement
	mp.priorityIndex.Set(tk, tx)

	return nil
}

// Select returns a set of transactions from the mempool, ordered by priority and sender-nonce in O(n) time.
// The passed in list of transactions are ignored.  This is a readonly operation, the mempool is not modified.
// maxBytes is the maximum number of bytes of transactions to return.
func (mp *priorityMempool) Select(_ [][]byte, maxBytes int64) ([]Tx, error) {
	var selectedTxs []Tx
	var txBytes int64
	mp.senderCursors = make(map[string]*huandu.Element)

	priorityNode := mp.priorityIndex.Front()
	for priorityNode != nil {
		priorityKey := priorityNode.Key().(txMeta)
		nextHighestPriority, nextPriorityNode := mp.nextPriority(priorityNode)
		sender := priorityKey.sender
		senderTx := mp.fetchSenderCursor(sender)

		// iterate through the sender's transactions in nonce order
		for senderTx != nil {
			mempoolTx := senderTx.Value.(Tx)
			if mp.onRead != nil {
				mp.onRead(mempoolTx)
			}

			key := senderTx.Key().(txMeta)

			fmt.Printf("read key: %v\n", key)

			// break if we've reached a transaction with a priority lower than the next highest priority in the pool
			if key.priority < nextHighestPriority {
				break
			} else if key.priority == nextHighestPriority {
				weight := mp.scores[txMeta{nonce: key.nonce, sender: key.sender}].weight
				if weight < nextPriorityNode.Key().(txMeta).weight {
					fmt.Printf("skipping %v due to weight %v\n", key, weight)
					break
				}
			}

			// otherwise, select the transaction and continue iteration
			selectedTxs = append(selectedTxs, mempoolTx)
			if txBytes += mempoolTx.Size(); txBytes >= maxBytes {
				return selectedTxs, nil
			}

			senderTx = senderTx.Next()
			mp.senderCursors[sender] = senderTx
		}

		priorityNode = nextPriorityNode
	}

	return selectedTxs, nil
}

func senderWeight(senderCursor *huandu.Element) int64 {
	if senderCursor == nil {
		return 0
	}
	weight := senderCursor.Key().(txMeta).priority
	senderCursor = senderCursor.Next()
	for senderCursor != nil {
		p := senderCursor.Key().(txMeta).priority
		if p != weight {
			weight = p
		}
		senderCursor = senderCursor.Next()
	}

	return weight
}

func (mp *priorityMempool) fetchSenderCursor(sender string) *huandu.Element {
	senderTx, ok := mp.senderCursors[sender]
	if !ok {
		senderTx = mp.senderIndices[sender].Front()
	}
	return senderTx
}

func (mp *priorityMempool) nextPriority(priorityNode *huandu.Element) (int64, *huandu.Element) {
	var nextPriorityNode *huandu.Element
	if priorityNode == nil {
		nextPriorityNode = mp.priorityIndex.Front()
	} else {
		nextPriorityNode = priorityNode.Next()
	}

	if nextPriorityNode == nil {
		return math.MinInt64, nil
	}

	np := nextPriorityNode.Key().(txMeta).priority

	return np, nextPriorityNode
}

// CountTx returns the number of transactions in the mempool.
func (mp *priorityMempool) CountTx() int {
	return mp.priorityIndex.Len()
}

// Remove removes a transaction from the mempool in O(log n) time, returning an error if unsuccessful.
func (mp *priorityMempool) Remove(tx Tx) error {
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

	sk := txMeta{nonce: nonce, sender: sender}
	score, ok := mp.scores[sk]
	if !ok {
		return ErrTxNotFound
	}
	tk := txMeta{nonce: nonce, priority: score.priority, sender: sender, weight: score.weight}

	senderTxs, ok := mp.senderIndices[sender]
	if !ok {
		return fmt.Errorf("sender %s not found", sender)
	}

	mp.priorityIndex.Remove(tk)
	senderTxs.Remove(tk)
	delete(mp.scores, sk)
	mp.priorityCounts[score.priority] = mp.priorityCounts[score.priority] - 1

	return nil
}

func IsEmpty(mempool Mempool) error {
	mp := mempool.(*priorityMempool)
	if mp.priorityIndex.Len() != 0 {
		return fmt.Errorf("priorityIndex not empty")
	}
	for key, v := range mp.priorityCounts {
		if v != 0 {
			return fmt.Errorf("priorityCounts not zero at %v, got %v", key, v)
		}
	}
	for key, v := range mp.senderIndices {
		if v.Len() != 0 {
			return fmt.Errorf("senderIndex not empty for sender %v", key)
		}
	}
	return nil
}

func DebugPrintKeys(mempool Mempool) {
	mp := mempool.(*priorityMempool)
	n := mp.priorityIndex.Front()
	for n != nil {
		k := n.Key().(txMeta)
		fmt.Printf("%s, %d, %d, %d\n", k.sender, k.priority, k.nonce, k.weight)
		n = n.Next()
	}
}
