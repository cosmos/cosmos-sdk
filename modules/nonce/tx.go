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
	basecoin.TxMapper.RegisterImplementation(&Tx{}, TypeNonce, ByteNonce)
}

// Tx - XXX fill in
type Tx struct {
	Tx       basecoin.Tx `json:p"tx"`
	Sequence uint32
	Signers  []basecoin.Actor // or simple []data.Bytes (they are only pubkeys...)
	seqKey   []byte           //key to store the sequence number
}

var _ basecoin.TxInner = &Tx{}

// NewTx wraps the tx with a signable nonce
func NewTx(tx basecoin.Tx, sequence uint32, signers []basecoin.Actor) basecoin.Tx {

	//Generate the sequence key as the hash of the list of signers, sorted by address
	sort.Sort(basecoin.ByAddress(signers))
	seqKey := merkle.SimpleHashFromBinary(signers)

	return (Tx{
		Tx:       tx,
		Sequence: sequence,
		Signers:  signers,
		seqKey:   seqKey,
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

	// check the current state
	cur, err := getSeq(store, n.seqKey)
	if err != nil {
		return err
	}
	if n.Sequence != cur+1 {
		return errors.ErrBadNonce()
	}

	// make sure they all signed
	for _, s := range n.Signers {
		if !ctx.HasPermission(s) {
			return errors.ErrNotMember()
		}
	}

	//finally increment the sequence by 1
	err = setSeq(store, n.seqKey, cur+1)
	if err != nil {
		return err
	}

	return nil
}
