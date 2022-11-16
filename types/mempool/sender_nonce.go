package mempool

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	huandu "github.com/huandu/skiplist"
)

var _ Mempool = (*senderNonceMempool)(nil) // _ Iterator = (*nonceMempoolIterator)(nil)

type senderTxs struct {
	txQueue *huandu.SkipList
	head    *huandu.Element
}

func newSenderTxs() senderTxs {
	return senderTxs{
		head:    nil,
		txQueue: huandu.New(huandu.LessThanFunc(txKeyLessNonce)),
	}
}

func (s senderTxs) insert(key txKey, tx sdk.Tx) {
	s.txQueue.Set(key, tx)
	s.head = s.txQueue.Front()
}

func (s senderTxs) getMove() *huandu.Element {
	if s.head == nil {
		return nil
	}
	currentHead := s.head
	s.head = s.head.Next()
	return currentHead
}

func (s senderTxs) remove(key txKey) error {
	res := s.txQueue.Remove(key)
	if res == nil {
		return ErrTxNotFound
	}
	if s.head == res {
		s.head = s.txQueue.Front()
	}
	return nil
}

type senderNonceMempool struct {
	senders map[string]senderTxs
	txCount int
}

func NewSenderNonceMempool() Mempool {
	senderMap := make(map[string]senderTxs)
	snp := &senderNonceMempool{
		senders: senderMap,
		txCount: 0,
	}
	return snp
}

func (snm senderNonceMempool) Insert(_ sdk.Context, tx sdk.Tx) error {
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
	senderTxs, found := snm.senders[sender]
	if !found {
		senderTxs = newSenderTxs()
	}

	senderTxs.insert(tk, tx)
	snm.senders[sender] = senderTxs
	snm.txCount = snm.txCount + 1
	return nil
}

func (snm senderNonceMempool) Select(context sdk.Context, i [][]byte) Iterator {
	iter := &senderNonceMepoolIterator{
		mempool: &snm,
	}
	iter.Next()
	return iter
}

// CountTx returns the total count of txs in the mempool.
func (snm senderNonceMempool) CountTx() int {
	return snm.txCount
}

// Remove removes a tx from the mempool. It returns an error if the tx does not have at least one signer or the tx
// was not found in the pool.
func (snm senderNonceMempool) Remove(tx sdk.Tx) error {
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
	senderTxs, found := snm.senders[sender]
	if !found {
		return ErrTxNotFound
	}
	err = senderTxs.remove(tk)
	if err != nil {
		return err
	}

	snm.senders[sender] = senderTxs
	snm.txCount = snm.txCount - 1
	return nil
}

type senderNonceMepoolIterator struct {
	mempool   *senderNonceMempool
	currentTx *huandu.Element
}

func (i senderNonceMepoolIterator) Next() Iterator {
	for sender := range i.mempool.senders {
		senderTxs, found := i.mempool.senders[sender]
		if !found {
			continue
		}
		tx := senderTxs.getMove()
		if tx == nil {
			continue
		}
		return senderNonceMepoolIterator{
			currentTx: tx,
			mempool:   i.mempool,
		}
	}

	return nil
}

func (i senderNonceMepoolIterator) Tx() sdk.Tx {
	return i.currentTx.Value.(sdk.Tx)
}
