package mempool

import (
	"fmt"
	"math/rand"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	huandu "github.com/huandu/skiplist"
)

var (
	_ Mempool  = (*senderNonceMempool)(nil)
	_ Iterator = (*nonceMempoolIterator)(nil)
)

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

func (s *senderTxs) insert(key txKey, tx sdk.Tx) {
	s.txQueue.Set(key, tx)
	s.head = s.txQueue.Front()
}

func (s *senderTxs) getMove() *huandu.Element {
	if s.head == nil {
		return nil
	}
	currentHead := s.head
	s.head = s.head.Next()
	return currentHead
}

func (s *senderTxs) remove(key txKey) error {
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
	senders map[string]*senderTxs
	txCount int
	rnd     *rand.Rand
}

func NewSenderNonceMempool() Mempool {
	senderMap := make(map[string]*senderTxs)
	snp := &senderNonceMempool{
		senders: senderMap,
		txCount: 0,
	}
	snp.setSeed(time.Now().UnixNano())
	return snp
}

func NewSenderNonceMempoolWithSeed(seed int64) Mempool {
	senderMap := make(map[string]*senderTxs)
	snp := &senderNonceMempool{
		senders: senderMap,
		txCount: 0,
	}
	snp.setSeed(seed)
	return snp
}

func (snm *senderNonceMempool) setSeed(seed int64) {
	s1 := rand.NewSource(seed)
	snm.rnd = rand.New(s1)
}

func (snm *senderNonceMempool) Insert(_ sdk.Context, tx sdk.Tx) error {
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
	senderTxs, found := snm.senders[sender]
	if !found {
		newSenderTx := newSenderTxs()
		senderTxs = &newSenderTx
	}
	tk := txKey{nonce: nonce, sender: sender}
	senderTxs.insert(tk, tx)
	snm.senders[sender] = senderTxs
	snm.txCount = snm.txCount + 1
	return nil
}

func (snm *senderNonceMempool) Select(context sdk.Context, i [][]byte) Iterator {
	var senders []string
	for key := range snm.senders {
		senders = append(senders, key)
	}
	iter := &senderNonceMepoolIterator{
		mempool: snm,
		senders: senders,
	}

	newIter := iter.Next()
	if newIter == nil {
		return nil
	}
	return newIter
}

// CountTx returns the total count of txs in the mempool.
func (snm *senderNonceMempool) CountTx() int {
	return snm.txCount
}

// Remove removes a tx from the mempool. It returns an error if the tx does not have at least one signer or the tx
// was not found in the pool.
func (snm *senderNonceMempool) Remove(tx sdk.Tx) error {
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
	if senderTxs.txQueue.Len() == 0 {
		delete(snm.senders, sender)
	} else {
		snm.senders[sender] = senderTxs
	}
	snm.txCount = snm.txCount - 1
	return nil
}

type senderNonceMepoolIterator struct {
	mempool   *senderNonceMempool
	currentTx *huandu.Element
	senders   []string
	seed      int
}

func (i *senderNonceMepoolIterator) Next() Iterator {
	for len(i.senders) > 0 {
		senderIndex := i.mempool.rnd.Intn(len(i.senders))
		fmt.Println(senderIndex)
		sender := i.senders[senderIndex]
		senderTxs, found := i.mempool.senders[sender]
		if !found {
			i.senders = removeAtIndex(i.senders, senderIndex)
			continue
		}
		tx := senderTxs.getMove()
		if tx == nil {
			i.senders = removeAtIndex(i.senders, senderIndex)
			continue
		}
		return &senderNonceMepoolIterator{
			senders:   i.senders,
			currentTx: tx,
			mempool:   i.mempool,
		}
	}

	return nil
}

func (i *senderNonceMepoolIterator) Tx() sdk.Tx {
	return i.currentTx.Value.(sdk.Tx)
}

func removeAtIndex[T any](slice []T, index int) []T {
	return append(slice[:index], slice[index+1:]...)
}
