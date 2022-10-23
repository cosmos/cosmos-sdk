package mempool

import (
	"fmt"

	huandu "github.com/huandu/skiplist"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
)

var _ Mempool = (*senderPriorityMempool)(nil)

type senderPriorityMempool struct {
	priorityIndex *huandu.SkipList
	senderIndices map[string]*huandu.SkipList
	priorities    map[senderPriorityMetadata]int64
}

func NewSenderPriorityMempool() Mempool {
	return &senderPriorityMempool{
		priorityIndex: huandu.New(huandu.LessThanFunc(func(a, b any) int {
			keyA := a.(senderPriorityMetadata)
			keyB := b.(senderPriorityMetadata)
			res := huandu.Int64.Compare(keyA.priority, keyB.priority)
			if res != 0 {
				return res
			}
			res = huandu.String.Compare(keyA.sender, keyB.sender)
			if res != 0 {
				return res
			}

			return huandu.Uint64.Compare(keyA.nonce, keyB.nonce)
		})),
		senderIndices: make(map[string]*huandu.SkipList),
		priorities:    make(map[senderPriorityMetadata]int64),
	}
}

type senderPriorityMetadata struct {
	priority int64
	sender   string
	nonce    uint64
}

func (mp *senderPriorityMempool) Insert(ctx sdk.Context, tx Tx) error {
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

	key := senderPriorityMetadata{
		priority: priority,
		sender:   sender,
		nonce:    nonce,
	}

	senderIndex, ok := mp.senderIndices[sender]
	if !ok {
		senderIndex = huandu.New(huandu.Uint64)
		mp.senderIndices[sender] = senderIndex
	}

	senderIndex.Set(nonce, tx)

	priorityKey := senderPriorityMetadata{sender: sender, nonce: nonce}
	if oldPriority, txExists := mp.priorities[priorityKey]; txExists {
		mp.priorityIndex.Remove(senderPriorityMetadata{priority: oldPriority, sender: sender, nonce: nonce})
	}

	mp.priorityIndex.Set(key, true)
	mp.priorities[priorityKey] = priority

	return nil
}

func (mp *senderPriorityMempool) Select(_ [][]byte, maxBytes int64) ([]Tx, error) {
	var (
		selectedTxs []Tx
		txBytes     int64
	)

	seenSenders := make(map[string]bool)

	// iterate over priority index and prioritize txs from senders with the highest priority
	priorityNode := mp.priorityIndex.Front()
	for priorityNode != nil {
		key := priorityNode.Key().(senderPriorityMetadata)

		// skip senders which have already been selected
		if seenSenders[key.sender] {
			priorityNode = priorityNode.Next()
			continue
		}

		// process each sender's txs in order of nonce completely before moving on to the next sender
		senderIndex := mp.senderIndices[key.sender]
		senderNode := senderIndex.Front()
		for senderNode != nil {
			tx := senderNode.Value.(Tx)
			selectedTxs = append(selectedTxs, tx)
			txBytes += tx.Size()

			if txBytes >= maxBytes {
				break
			}

			senderNode = senderNode.Next()
		}

		if txBytes >= maxBytes {
			break
		}

		seenSenders[key.sender] = true

		priorityNode = priorityNode.Next()
	}

	return selectedTxs, nil
}

func (mp *senderPriorityMempool) CountTx() int {
	return mp.priorityIndex.Len()
}

func (mp *senderPriorityMempool) Remove(tx Tx) error {
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

	priority, ok := mp.priorities[senderPriorityMetadata{sender: sender, nonce: nonce}]
	if !ok {
		return ErrTxNotFound
	}

	mp.priorityIndex.Remove(senderPriorityMetadata{sender: sender, priority: priority})
	senderIndex, ok := mp.senderIndices[sender]
	if !ok {
		return fmt.Errorf("sender %s not found", sender)
	}
	senderIndex.Remove(nonce)

	return nil
}
