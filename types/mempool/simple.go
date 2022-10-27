package mempool

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	huandu "github.com/huandu/skiplist"
)

type simpleMempool struct {
	txQueue *huandu.SkipList
}

// type txKey struct {
//	nonce  uint64
//	sender string
//}
//
//func txKeyLessNonce(a, b any) int {
//	keyA := a.(txKey)
//	keyB := b.(txKey)
//
//	return huandu.Uint64.Compare(keyB.nonce, keyA.nonce)
//}

func DefaultSimpleMempool() Mempool {
	return NewPriorityMempool()
}

func NewSimpleMempool(opts ...PriorityMempoolOption) Mempool {
	sp := &simpleMempool{
		txQueue: huandu.New(huandu.LessThanFunc(txKeyLessNonce)),
	}

	return sp
}

func (sp simpleMempool) Insert(_ sdk.Context, tx Tx) error {
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
	fmt.Println("key:", tk)
	sp.txQueue.Set(tk, tx)
	fmt.Println("length of queue", sp.CountTx())
	return nil
}

func (sp simpleMempool) Select(txs [][]byte, maxBytes int64) ([]Tx, error) {
	var selectedTxs []Tx

	currentTx := sp.txQueue.Front()
	for currentTx != nil {
		mempoolTx := currentTx.Value.(Tx)

		selectedTxs = append(selectedTxs, mempoolTx)
		// if txBytes += mempoolTx.Size(); txBytes >= maxBytes {
		//	return selectedTxs, nil
		//}
		currentTx = currentTx.Next()
	}
	return selectedTxs, nil
}

func (sp simpleMempool) CountTx() int {
	return sp.txQueue.Len()
}

func (sp simpleMempool) Remove(tx Tx) error {
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
	sp.txQueue.Remove(tk)
	return nil
}
