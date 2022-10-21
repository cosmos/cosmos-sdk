package mempool

import (
	"fmt"
	"math"

	huandu "github.com/huandu/skiplist"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
)

var _ Mempool = (*priorityMempool)(nil)

// priorityMempool defines the SDK's default mempool implementation which stores
// txs in a partially ordered set by 2 dimensions: priority, and sender-nonce
// (sequence number). Internally it uses one priority ordered skip list and one
// skip list per sender ordered by sender-nonce (sequence number). When there
// are multiple txs from the same sender, they are not always comparable by
// priority to other sender txs and must be partially ordered by both sender-nonce
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
	// nonce is the sender's sequence number
	nonce uint64
	// priority is the transaction's priority
	priority int64
	// sender is the transaction's sender
	sender string
	// weight is the transaction's weight, used as a tiebreaker for transactions with the same priority
	weight int64
	// senderElement is a pointer to the transaction's element in the sender index
	senderElement *huandu.Element
}

// txMetaLess is a comparator for txKeys that first compares priority, then weight,
// then sender, then nonce, uniquely identifying a transaction.
//
// Note, txMetaLess is used as the comparator in the priority index.
func txMetaLess(a, b any) int {
	keyA := a.(txMeta)
	keyB := b.(txMeta)
	res := huandu.Int64.Compare(keyA.priority, keyB.priority)
	if res != 0 {
		return res
	}

	// weight is used as a tiebreaker for transactions with the same priority.  weight is calculated in a single
	// pass in .Select(...) and so will be 0 on .Insert(...)
	res = huandu.Int64.Compare(keyA.weight, keyB.weight)
	if res != 0 {
		return res
	}

	// Because weight will be 0 on .Insert(...), we must also compare sender and nonce to resolve priority collisions.
	// If we didn't then transactions with the same priority would overwrite each other in the priority index.
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

// NewPriorityMempool returns the SDK's default mempool implementation which
// returns txs in a partial order by 2 dimensions; priority, and sender-nonce.
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

// Insert attempts to insert a Tx into the app-side mempool in O(log n) time,
// returning an error if unsuccessful. Sender and nonce are derived from the
// transaction's first signature.
//
// Transactions are unique by sender and nonce. Inserting a duplicate tx is an
// O(log n) no-op.
//
// Inserting a duplicate tx with a different priority overwrites the existing tx,
// changing the total order of the mempool.
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
	key := txMeta{nonce: nonce, priority: priority, sender: sender}

	senderIndex, ok := mp.senderIndices[sender]
	if !ok {
		senderIndex = huandu.New(huandu.LessThanFunc(func(a, b any) int {
			return huandu.Uint64.Compare(b.(txMeta).nonce, a.(txMeta).nonce)
		}))

		// initialize sender index if not found
		mp.senderIndices[sender] = senderIndex
	}

	mp.priorityCounts[priority]++

	// Since senderIndex is scored by nonce, a changed priority will overwrite the
	// existing key.
	key.senderElement = senderIndex.Set(key, tx)

	// Since mp.priorityIndex is scored by priority, then sender, then nonce, a
	// changed priority will create a new key, so we must remove the old key and
	// re-insert it to avoid having the same tx with different priorityIndex indexed
	// twice in the mempool.
	//
	// This O(log n) remove operation is rare and only happens when a tx's priority
	// changes.
	sk := txMeta{nonce: nonce, sender: sender}
	if oldScore, txExists := mp.scores[sk]; txExists {
		mp.priorityIndex.Remove(txMeta{
			nonce:    nonce,
			sender:   sender,
			priority: oldScore.priority,
			weight:   oldScore.weight,
		})
	}

	mp.scores[sk] = txMeta{priority: priority}
	mp.priorityIndex.Set(key, tx)

	return nil
}

// Select returns a set of transactions from the mempool, ordered by priority
// and sender-nonce in O(n) time. The passed in list of transactions are ignored.
// This is a readonly operation, the mempool is not modified.
//
// The maxBytes parameter defines the maximum number of bytes of transactions to
// return.
func (mp *priorityMempool) Select(_ [][]byte, maxBytes int64) ([]Tx, error) {
	var (
		selectedTxs []Tx
		txBytes     int64
	)

	mp.senderCursors = make(map[string]*huandu.Element)
	mp.reorderPriorityTies()

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

			// break if we've reached a transaction with a priority lower than the next highest priority in the pool
			if key.priority < nextHighestPriority {
				break
			} else if key.priority == nextHighestPriority {
				// weight is incorporated into the priority index key only (not sender index) so we must fetch it here
				// from the scores map.
				weight := mp.scores[txMeta{nonce: key.nonce, sender: key.sender}].weight
				if weight < nextPriorityNode.Key().(txMeta).weight {
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

type reorderKey struct {
	deleteKey txMeta
	insertKey txMeta
	tx        Tx
}

func (mp *priorityMempool) reorderPriorityTies() {
	node := mp.priorityIndex.Front()
	var reordering []reorderKey
	for node != nil {
		key := node.Key().(txMeta)
		if mp.priorityCounts[key.priority] > 1 {
			newKey := key
			newKey.weight = senderWeight(key.senderElement)
			reordering = append(reordering, reorderKey{deleteKey: key, insertKey: newKey, tx: node.Value.(Tx)})
		}
		node = node.Next()
	}

	for _, k := range reordering {
		mp.priorityIndex.Remove(k.deleteKey)
		delete(mp.scores, txMeta{nonce: k.deleteKey.nonce, sender: k.deleteKey.sender})
		mp.priorityIndex.Set(k.insertKey, k.tx)
		mp.scores[txMeta{nonce: k.insertKey.nonce, sender: k.insertKey.sender}] = k.insertKey
	}
}

// senderWeight returns the weight of a given tx (t) at senderCursor.  Weight is defined as the first (nonce-wise)
// same sender tx with a priority not equal to t.  It is used to resolve priority collisions, that is when 2 or more
// txs from different senders have the same priority.
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

	scoreKey := txMeta{nonce: nonce, sender: sender}
	score, ok := mp.scores[scoreKey]
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
	delete(mp.scores, scoreKey)
	mp.priorityCounts[score.priority]--

	return nil
}

func IsEmpty(mempool Mempool) error {
	mp := mempool.(*priorityMempool)
	if mp.priorityIndex.Len() != 0 {
		return fmt.Errorf("priorityIndex not empty")
	}

	var countKeys []int64
	for k := range mp.priorityCounts {
		countKeys = append(countKeys, k)
	}
	for _, k := range countKeys {
		if mp.priorityCounts[k] != 0 {
			return fmt.Errorf("priorityCounts not zero at %v, got %v", k, mp.priorityCounts[k])
		}
	}

	var senderKeys []string
	for k := range mp.senderIndices {
		senderKeys = append(senderKeys, k)
	}
	for _, k := range senderKeys {
		if mp.senderIndices[k].Len() != 0 {
			return fmt.Errorf("senderIndex not empty for sender %v", k)
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
