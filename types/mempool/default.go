package mempool

import (
	"fmt"
	"math"

	huandu "github.com/huandu/skiplist"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
)

var _ Mempool = (*defaultMempool)(nil)

// defaultMempool is the SDK's default mempool implementation which stores txs in a partially ordered set
// by 2 dimensions: priority, and sender-nonce (sequence number).  Internally it uses one priority ordered skip list
// and one skip per sender ordered by nonce (sender-nonce).  When there are multiple txs from the same sender,
// they are not always comparable by priority to other sender txs and must be partially ordered by both sender-nonce
// and priority.
type defaultMempool struct {
	priorityIndex *huandu.SkipList
	senderIndices map[string]*huandu.SkipList
	scores        map[txKey]int64
	onRead        func(tx Tx)
}

type txKey struct {
	nonce    uint64
	priority int64
	sender   string
}

// txKeyLess is a comparator for txKeys that first compares priority, then sender, then nonce, uniquely identifying
// a transaction.
func txKeyLess(a, b any) int {
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

type DefaultMempoolOption func(*defaultMempool)

// WithOnRead sets a callback to be called when a tx is read from the mempool.
func WithOnRead(onRead func(tx Tx)) DefaultMempoolOption {
	return func(mp *defaultMempool) {
		mp.onRead = onRead
	}
}

// NewDefaultMempool returns the SDK's default mempool implementation which returns txs in a partial order
// by 2 dimensions; priority, and sender-nonce.
func NewDefaultMempool(opts ...DefaultMempoolOption) Mempool {
	mp := &defaultMempool{
		priorityIndex: huandu.New(huandu.LessThanFunc(txKeyLess)),
		senderIndices: make(map[string]*huandu.SkipList),
		scores:        make(map[txKey]int64),
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
func (mp *defaultMempool) Insert(ctx sdk.Context, tx Tx) error {
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

	senderIndex, ok := mp.senderIndices[sender]
	if !ok {
		senderIndex = huandu.New(huandu.LessThanFunc(func(a, b any) int {
			return huandu.Uint64.Compare(b.(txKey).nonce, a.(txKey).nonce)
		}))
		// initialize sender index if not found
		mp.senderIndices[sender] = senderIndex
	}

	// Since senderIndex is scored by nonce, a changed priority will overwrite the existing txKey.
	senderIndex.Set(tk, tx)

	// Since mp.priorityIndex is scored by priority, then sender, then nonce, a changed priority will create a new key,
	// so we must remove the old key and re-insert it to avoid having the same tx with different priorityIndex indexed
	// twice in the mempool.  This O(log n) remove operation is rare and only happens when a tx's priority changes.
	sk := txKey{nonce: nonce, sender: sender}
	if oldScore, txExists := mp.scores[sk]; txExists {
		mp.priorityIndex.Remove(txKey{nonce: nonce, priority: oldScore, sender: sender})
	}
	mp.scores[sk] = ctx.Priority()
	mp.priorityIndex.Set(tk, tx)

	return nil
}

// Select returns a set of transactions from the mempool, ordered by priority and sender-nonce in O(n) time.
// The passed in list of transactions are ignored.  This is a readonly operation, the mempool is not modified.
// maxBytes is the maximum number of bytes of transactions to return.
func (mp *defaultMempool) Select(_ [][]byte, maxBytes int64) ([]Tx, error) {
	var selectedTxs []Tx
	var txBytes int64
	senderCursors := make(map[string]*huandu.Element)

	// start with the highest priority sender
	priorityNode := mp.priorityIndex.Front()
	for priorityNode != nil {
		priorityKey := priorityNode.Key().(txKey)
		nextHighestPriority, nextPriorityNode := nextPriority(priorityNode)
		sender := priorityKey.sender
		senderTx := mp.fetchSenderCursor(senderCursors, sender)

		// iterate through the sender's transactions in nonce order
		for senderTx != nil {
			mempoolTx := senderTx.Value.(Tx)
			if mp.onRead != nil {
				mp.onRead(mempoolTx)
			}

			k := senderTx.Key().(txKey)

			// break if we've reached a transaction with a priority lower than the next highest priority in the pool
			if k.priority < nextHighestPriority {
				break
			}

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
		senderTx = mp.senderIndices[sender].Front()
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
	return mp.priorityIndex.Len()
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
		return ErrTxNotFound
	}
	tk := txKey{nonce: nonce, priority: priority, sender: sender}

	senderTxs, ok := mp.senderIndices[sender]
	if !ok {
		return fmt.Errorf("sender %s not found", sender)
	}

	mp.priorityIndex.Remove(tk)
	senderTxs.Remove(tk)
	delete(mp.scores, sk)

	return nil
}

func DebugPrintKeys(mempool Mempool) {
	mp := mempool.(*defaultMempool)
	n := mp.priorityIndex.Front()
	for n != nil {
		k := n.Key().(txKey)
		fmt.Printf("%s, %d, %d\n", k.sender, k.priority, k.nonce)
		n = n.Next()
	}
}
