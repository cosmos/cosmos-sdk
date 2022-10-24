package mempool

import (
	"fmt"

	huandu "github.com/huandu/skiplist"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
)

type nonceMempool struct {
	txQueue *huandu.SkipList
}

type txKey struct {
	nonce  uint64
	sender string
}

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

func (sp nonceMempool) Select(txs [][]byte, maxBytes int64) ([]Tx, error) {
	var (
		txBytes     int64
		selectedTxs []Tx
	)

	currentTx := sp.txQueue.Front()
	for currentTx != nil {
		mempoolTx := currentTx.Value.(Tx)

		selectedTxs = append(selectedTxs, mempoolTx)
		if txBytes += mempoolTx.Size(); txBytes >= maxBytes {
			return selectedTxs, nil
		}
		currentTx = currentTx.Next()
	}
	return selectedTxs, nil
}

func (sp nonceMempool) CountTx() int {
	return sp.txQueue.Len()
}

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
