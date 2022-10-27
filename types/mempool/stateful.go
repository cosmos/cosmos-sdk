package mempool

import (
	"bytes"
	"crypto/sha256"
	"fmt"

	huandu "github.com/huandu/skiplist"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
)

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
func (amp *AccountMemPool) Push(ctx sdk.Context, key statefullPriorityKey, tx Tx) {
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

func (amp *MemPoolI) Insert(ctx sdk.Context, tx Tx) error {
	senders := tx.(signing.SigVerifiableTx).GetSigners()
	nonces, err := tx.(signing.SigVerifiableTx).GetSignaturesV2()

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
	hash := sha256.Sum256(senders[0].Bytes())
	key := statefullPriorityKey{hash: hash, nonce: nonce, priority: ctx.Priority()}

	prevKey := accountMeempool.currentKey
	accountMeempool.Push(ctx, key, tx)

	amp.accountsHeads.Remove(prevKey)
	amp.accountsHeads.Set(accountMeempool.currentKey, accountMeempool)
	amp.senders[sender] = accountMeempool
	return nil
}

func (amp *MemPoolI) Select(_ sdk.Context, _ [][]byte, maxBytes int64) ([]Tx, error) {
	var selectedTxs []Tx
	var txBytes int64

	currentAccount := amp.accountsHeads.Front()
	for currentAccount != nil {
		accountMemPool := currentAccount.Value.(*AccountMemPool)

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
