package mempool

import (
	"fmt"

	huandu "github.com/huandu/skiplist"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
)

var (
	_ Mempool  = (*nonceMempool)(nil)
	_ Iterator = (*nonceMempoolIterator)(nil)
)

// nonceMempool is a mempool that keeps transactions sorted by nonce. Transactions
// with the lowest nonce globally are prioritized. Transactions with the same
// nonce are prioritized by sender address. Fee/gas based prioritization is not
// supported.
type nonceMempool struct {
	txQueue *huandu.SkipList
}

type nonceMempoolIterator struct {
	currentTx *huandu.Element
}

func (i nonceMempoolIterator) Next() Iterator {
	if i.currentTx == nil {
		return nil
	} else if n := i.currentTx.Next(); n != nil {
		return nonceMempoolIterator{currentTx: n}
	} else {
		return nil
	}
}

func (i nonceMempoolIterator) Tx() sdk.Tx {
	return i.currentTx.Value.(sdk.Tx)
}

type txKey struct {
	nonce  uint64
	sender string
}

// txKeyLessNonce compares two txKeys by nonce then by sender address.
func txKeyLessNonce(a, b any) int {
	keyA := a.(txKey)
	keyB := b.(txKey)

	res := huandu.Uint64.Compare(keyB.nonce, keyA.nonce)
	if res != 0 {
		return res
	}

	return huandu.String.Compare(keyB.sender, keyA.sender)
}

// NewNonceMempool creates a new mempool that prioritizes transactions by nonce, the lowest first.
func NewNonceMempool() Mempool {
	sp := &nonceMempool{
		txQueue: huandu.New(huandu.LessThanFunc(txKeyLessNonce)),
	}

	return sp
}

// Insert adds a tx to the mempool. It returns an error if the tx does not have at least one signer.
// priority is ignored.
func (sp nonceMempool) Insert(_ sdk.Context, tx sdk.Tx) error {
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
	tk := txKey{nonce: nonce, sender: sender}
	sp.txQueue.Set(tk, tx)
	return nil
}

// Select returns an iterator ordering transactions the mempool with the lowest nonce globally first. A sender's txs
// will always be returned in nonce order.
func (sp nonceMempool) Select(_ sdk.Context, _ [][]byte) Iterator {
	currentTx := sp.txQueue.Front()
	if currentTx == nil {
		return nil
	}

	return &nonceMempoolIterator{currentTx: currentTx}
}

// CountTx returns the number of txs in the mempool.
func (sp nonceMempool) CountTx() int {
	return sp.txQueue.Len()
}

// Remove removes a tx from the mempool. It returns an error if the tx does not have at least one signer or the tx
// was not found in the pool.
func (sp nonceMempool) Remove(tx sdk.Tx) error {
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
	tk := txKey{nonce: nonce, sender: sender}
	res := sp.txQueue.Remove(tk)
	if res == nil {
		return ErrTxNotFound
	}
	return nil
}
