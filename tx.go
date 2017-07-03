package basecoin

import (
	"github.com/pkg/errors"
	"github.com/tendermint/go-wire/data"
)

const maxTxSize = 10240

// TxInner is the interface all concrete transactions should implement.
//
// It adds bindings for clean un/marhsaling of the various implementations
// both as json and binary, as well as some common functionality to move them.
//
// +gen wrapper:"Tx"
type TxInner interface {
	Wrap() Tx

	// ValidateBasic should be a stateless check and just verify that the
	// tx is properly formated (required strings not blank, signatures exist, etc.)
	// this can also be run on the client-side for better debugging before posting a tx
	ValidateBasic() error
}

// LoadTx parses a tx from data
//
// TODO: label both errors with abci.CodeType_EncodingError
// need to move errors to avoid import cycle
func LoadTx(bin []byte) (tx Tx, err error) {
	if len(bin) > maxTxSize {
		return tx, errors.New("Tx size exceeds maximum")
	}

	// Decode tx
	err = data.FromWire(bin, &tx)
	return tx, err
}

// TODO: do we need this abstraction? TxLayer???
// please review again after implementing "middleware"

// TxLayer provides a standard way to deal with "middleware" tx,
// That add context to an embedded tx.
type TxLayer interface {
	TxInner
	Next() Tx
}

func (t Tx) IsLayer() bool {
	_, ok := t.Unwrap().(TxLayer)
	return ok
}

func (t Tx) GetLayer() TxLayer {
	l, _ := t.Unwrap().(TxLayer)
	return l
}
