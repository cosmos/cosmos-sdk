package mempool

import (
	"fmt"

	huandu "github.com/huandu/skiplist"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
)

// nonceMempool is a mempool that keeps transactions sorted by nonce. Transactions with the lowest nonce globally
// are prioritized. Transactions with the same nonce are prioritized by sender address. Fee/gas based
// prioritization is not supported.
type nonceMempool struct {
	txQueue *huandu.SkipList
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

func NewNonceMempool() Mempool {
	sp := &nonceMempool{
		txQueue: huandu.New(huandu.LessThanFunc(txKeyLessNonce)),
	}

	return sp
}

// Insert adds a tx to the mempool. It returns an error if the tx does not have at least one signer.
// priority is ignored.
func (sp nonceMempool) Insert(_ sdk.Context, tx Tx) error {
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

// Select returns txs from the mempool with the lowest nonce globally first. A sender's txs will always be returned
// in nonce order.
func (sp nonceMempool) Select(_ [][]byte, maxBytes int64) ([]Tx, error) {
	var (
		txBytes     int64
		selectedTxs []Tx
	)

	currentTx := sp.txQueue.Front()
	for currentTx != nil {
		mempoolTx := currentTx.Value.(Tx)

		if txBytes += mempoolTx.Size(); txBytes <= maxBytes {
			selectedTxs = append(selectedTxs, mempoolTx)
		} else {
			return selectedTxs, nil
		}
		currentTx = currentTx.Next()
	}
	return selectedTxs, nil
}

// CountTx returns the number of txs in the mempool.
func (sp nonceMempool) CountTx() int {
	return sp.txQueue.Len()
}

// Remove removes a tx from the mempool. It returns an error if the tx does not have at least one signer or the tx
// was not found in the pool.
func (sp nonceMempool) Remove(tx Tx) error {
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
