/*
Package nonce XXX
*/
package nonce

import (
	"sort"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/errors"
	"github.com/tendermint/basecoin/state"

	"github.com/tendermint/tmlibs/merkle"
)

// nolint
const (
	ByteNonce = 0x69 //TODO overhaul byte assign system don't make no sense!
	TypeNonce = "nonce"
)

func init() {
	basecoin.TxMapper.RegisterImplementation(Tx{}, TypeNonce, ByteNonce)
}

// Tx - XXX fill in
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
	return n.Tx.ValidateBasic()
}

// CheckIncrementSeq - XXX fill in
func (n Tx) CheckIncrementSeq(ctx basecoin.Context, store state.KVStore) error {

	//Generate the sequence key as the hash of the list of signers, sorted by address
	sort.Sort(basecoin.ByAddress(n.Signers))
	seqKey := merkle.SimpleHashFromBinary(n.Signers)

	// check the current state
	cur, err := getSeq(store, seqKey)
	if err != nil {
		return err
	}
	if n.Sequence != cur+1 {
		return errors.ErrBadNonce(n.Sequence, cur+1)
	}

	// make sure they all signed
	for _, s := range n.Signers {
		if !ctx.HasPermission(s) {
			return errors.ErrNotMember()
		}
	}

	//finally increment the sequence by 1
	err = setSeq(store, seqKey, cur+1)
	if err != nil {
		return err
	}

	return nil
}
