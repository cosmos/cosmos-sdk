package mempool

import (
	crand "crypto/rand" // #nosec // crypto/rand is used for seed generation
	"encoding/binary"
	"fmt"
	"math/rand" // #nosec // math/rand is used for random selection and seeded from crypto/rand

	huandu "github.com/huandu/skiplist"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
)

var (
	_ Mempool  = (*senderNonceMempool)(nil)
	_ Iterator = (*senderNonceMepoolIterator)(nil)
)

type senderTxs struct {
	cursor *huandu.Element
}

func newSenderTxs(tx *huandu.Element) senderTxs {
	return senderTxs{
		cursor: tx,
	}
}

func (s *senderTxs) next() *huandu.Element {
	if s.cursor == nil {
		return nil
	}
	currentCursor := s.cursor
	s.cursor = s.cursor.Next()
	return currentCursor
}

type senderNonceMempool struct {
	senders map[string]*huandu.SkipList
	rnd     *rand.Rand
}

// NewSenderNonceMempool creates a new mempool that prioritizes transactions by nonce, the lowest first.
func NewSenderNonceMempool() Mempool {
	senderMap := make(map[string]*huandu.SkipList)
	snp := &senderNonceMempool{
		senders: senderMap,
	}

	var seed int64
	binary.Read(crand.Reader, binary.BigEndian, &seed)
	snp.setSeed(seed)

	return snp
}

// NewSenderNonceMempoolWithSeed creates a new mempool that prioritizes transactions by nonce, the lowest first and sets the random seed.
func NewSenderNonceMempoolWithSeed(seed int64) Mempool {
	senderMap := make(map[string]*huandu.SkipList)
	snp := &senderNonceMempool{
		senders: senderMap,
	}
	snp.setSeed(seed)
	return snp
}

func (snm *senderNonceMempool) setSeed(seed int64) {
	s1 := rand.NewSource(seed)
	snm.rnd = rand.New(s1) //#nosec // math/rand is seeded from crypto/rand by default
}

// Insert adds a tx to the mempool. It returns an error if the tx does not have at least one signer.
// priority is ignored.
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
		senderTxs = huandu.New(huandu.Uint64)
		snm.senders[sender] = senderTxs
	}
	senderTxs.Set(nonce, tx)

	return nil
}

// Select returns an iterator ordering transactions the mempool with the lowest nonce of a random selected sender first.
func (snm *senderNonceMempool) Select(_ sdk.Context, _ [][]byte) Iterator {
	var senders []string
	senderCursors := make(map[string]*senderTxs)
	// #nosec
	for key := range snm.senders {
		senders = append(senders, key)
		senderTx := newSenderTxs(snm.senders[key].Front())
		senderCursors[key] = &senderTx
	}

	iter := &senderNonceMepoolIterator{
		senders:         senders,
		rnd:             snm.rnd,
		sendersCurosors: senderCursors,
	}

	newIter := iter.Next()
	if newIter == nil {
		return nil
	}
	return newIter
}

// CountTx returns the total count of txs in the mempool.
func (snm *senderNonceMempool) CountTx() int {
	count := 0

	// Disable gosec here since we need neither strong randomness nor deterministic iteration.
	// #nosec
	for _, value := range snm.senders {
		count += value.Len()
	}
	return count
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
	senderTxs, found := snm.senders[sender]
	if !found {
		return ErrTxNotFound
	}

	res := senderTxs.Remove(nonce)
	if res == nil {
		return ErrTxNotFound
	}

	if senderTxs.Len() == 0 {
		delete(snm.senders, sender)
	}
	return nil
}

type senderNonceMepoolIterator struct {
	rnd             *rand.Rand
	currentTx       *huandu.Element
	senders         []string
	sendersCurosors map[string]*senderTxs
}

// Next it returns the iterator next state where a iterator will contain a tx that was the smallest
// nonce of a randomly selected sender
func (i *senderNonceMepoolIterator) Next() Iterator {
	for len(i.senders) > 0 {
		senderIndex := i.rnd.Intn(len(i.senders))
		sender := i.senders[senderIndex]
		senderTxs, found := i.sendersCurosors[sender]
		if !found {
			i.senders = removeAtIndex(i.senders, senderIndex)
			continue
		}
		tx := senderTxs.next()
		if tx == nil {
			i.senders = removeAtIndex(i.senders, senderIndex)
			continue
		}
		return &senderNonceMepoolIterator{
			senders:         i.senders,
			currentTx:       tx,
			rnd:             i.rnd,
			sendersCurosors: i.sendersCurosors,
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
