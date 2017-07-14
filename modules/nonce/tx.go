/*
Package nonce - This module allows replay protection to be added to process stack.
This is achieved through the use of a sequence number for each unique set of signers.
Note that the sequence number for the single signing account "foo" will be unique
from the sequence number for a multi-sig account {"foo", "bar"} which would also be
unique from a different multi-sig account {"foo", "soup"}
*/
package nonce

import (
	"sort"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/errors"
	"github.com/tendermint/basecoin/state"
)

// nolint
const (
	ByteNonce = 0x69 //TODO overhaul byte assign system don't make no sense!
	TypeNonce = "nonce"
)

func init() {
	basecoin.TxMapper.RegisterImplementation(Tx{}, TypeNonce, ByteNonce)
}

// Tx - Nonce transaction structure, contains list of signers and current sequence number
type Tx struct {
	Sequence uint32           `json:"sequence"`
	Signers  []basecoin.Actor `json:"signers"`
	Tx       basecoin.Tx      `json:"tx"`
}

var _ basecoin.TxInner = &Tx{}

// NewTx wraps the tx with a signable nonce
func NewTx(sequence uint32, signers []basecoin.Actor, tx basecoin.Tx) basecoin.Tx {
	return (Tx{
		Sequence: sequence,
		Signers:  signers,
		Tx:       tx,
	}).Wrap()
}

//nolint
func (n Tx) Wrap() basecoin.Tx {
	return basecoin.Tx{n}
}
func (n Tx) ValidateBasic() error {
	switch {
	case n.Tx.Empty():
		return errors.ErrTxEmpty()
	case n.Sequence == 0:
		return ErrZeroSequence()
	case len(n.Signers) == 0:
		return errors.ErrNoSigners()
	}
	return n.Tx.ValidateBasic()
}

// CheckIncrementSeq - Check that the sequence number is one more than the state sequence number
// and further increment the sequence number
// NOTE It is okay to modify the sequence before running the wrapped TX because if the
// wrapped Tx fails, the state changes are not applied
func (n Tx) CheckIncrementSeq(ctx basecoin.Context, store state.KVStore) error {

	seqKey := n.getSeqKey()

	// check the current state
	cur, err := getSeq(store, seqKey)
	if err != nil {
		return err
	}
	if n.Sequence != cur+1 {
		return ErrBadNonce(n.Sequence, cur+1)
	}

	// make sure they all signed
	for _, s := range n.Signers {
		if !ctx.HasPermission(s) {
			return ErrNotMember()
		}
	}

	// increment the sequence by 1
	err = setSeq(store, seqKey, cur+1)
	if err != nil {
		return err
	}

	return nil
}

// Generate the sequence key as the concatenated list of signers, sorted by address.
func (n Tx) getSeqKey() (seqKey []byte) {

	// First copy the list of signers to sort as sort is done in place
	signers2sort := make([]basecoin.Actor, len(n.Signers))
	copy(signers2sort, n.Signers)
	sort.Sort(basecoin.ByAddress(n.Signers))

	for _, signer := range n.Signers {
		seqKey = append(seqKey, signer.Address...)
	}
	//seqKey = merkle.SimpleHashFromBinary(n.Signers)
	return
}
