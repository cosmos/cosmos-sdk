/*
Package nonce XXX
*/
package nonce

import (
	"github.com/tendermint/basecoin/state"

	"github.com/tendermint/basecoin"
)

// nolint
const (
	// for signatures
	ByteSingleTx = 0x16
	ByteMultiSig = 0x17
)

/**** Registration ****/

//func init() {
//basecoin.TxMapper.RegisterImplementation(&Tx{}, TypeSingleTx, ByteSingleTx)
//}

// Tx - XXX fill in
type Tx struct {
	Tx       basecoin.Tx `json:p"tx"`
	Sequence uint32
	Signers  []basecoin.Actor // or simple []data.Bytes (they are only pubkeys...)
}

var _ basecoin.TxInner = &Tx{}

// NewTx wraps the tx with a signable nonce
func NewTx(tx basecoin.Tx, sequence uint32, signers []basecoin.Actor) *Tx {
	return &Tx{
		Tx:       tx,
		Sequence: sequence,
		Signers:  signers,
	}
}

//nolint
func (n *Tx) Wrap() basecoin.Tx {
	return basecoin.Tx{s}
}
func (n *Tx) ValidateBasic() error {
	return s.Tx.ValidateBasic()
}

// CheckSequence - XXX fill in
func (n *Tx) CheckSequence(ctx basecoin.Context, store state.KVStore) error {

	// check the current state
	h := hash(Sort(n.Signers))
	cur := loadSeq(store, h)
	if n.Sequence != cur+1 {
		return ErrBadNonce()
	}

	// make sure they all signed
	for _, s := range n.Signers {
		if !ctx.HasPermission(s) {
			return ErrNotMember()
		}
	}

	return nil
}
