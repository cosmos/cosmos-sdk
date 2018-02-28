package types

import (
	"encoding/json"

	crypto "github.com/tendermint/go-crypto"
)

// Transactions messages must fulfill the Msg
type Msg interface {

	// Return the message type.
	// Must be alphanumeric or empty.
	Type() string

	// Get some property of the Msg.
	Get(key interface{}) (value interface{})

	// Get the canonical byte representation of the Msg.
	GetSignBytes(Context) []byte

	// ValidateBasic does a simple validation check that
	// doesn't require access to any other information.
	ValidateBasic() Error

	// Signers returns the addrs of signers that must sign.
	// CONTRACT: All signatures must be present to be valid.
	// CONTRACT: Returns addrs in some deterministic order.
	GetSigners() []crypto.Address
}

// Transactions objects must fulfill the Tx
type Tx interface {

	// Gets the Msg.
	GetMsg() Msg

	// The address that pays the base fee for this message.  The fee is
	// deducted before the Msg is processed.
	GetFeePayer() crypto.Address

	// Signatures returns the signature of signers who signed the Msg.
	// CONTRACT: Length returned is same as length of
	// pubkeys returned from MsgKeySigners, and the order
	// matches.
	// CONTRACT: If the signature is missing (ie the Msg is
	// invalid), then the corresponding signature is
	// .Empty().
	GetSignatures() []StdSignature

	// AccNonce returns the nonce of the feepayer account
	GetAccNonce() int64

	// TxSequence returns the nonce of the tx within the feepayer account
	GetTxNonce() int64
}

var _ Tx = (*StdTx)(nil)

// StdTx is a standard way to wrap a Msg with Signatures.
// NOTE: the first signature is the FeePayer (Signatures must not be nil).
// NOTE: a single Sequence is included in the StdTx, separate from the Msg
// 	- it must be incldued in the SignBytes to actually provide replay protection
type StdTx struct {
	Msg
	Signatures []StdSignature
	TxNonce    int64
	AccNonce   int64
}

func NewStdTx(msg Msg, sigs []StdSignature) StdTx {
	return StdTx{
		Msg:        msg,
		Signatures: sigs,
	}
}

//nolint
func (tx StdTx) GetMsg() Msg                   { return tx.Msg }
func (tx StdTx) GetFeePayer() crypto.Address   { return tx.GetSigners()[0] }
func (tx StdTx) GetSignatures() []StdSignature { return tx.Signatures }
func (tx StdTx) GetAccNonce() int64            { return tx.AccNonce }
func (tx StdTx) GetTxNonce() int64             { return tx.TxNonce }

func CanonicalSignBytes(ctx Context, msg Msg) []byte {
	obj := SignBytesObject{
		ChainID:  "", // TODO
		AccNonce: 0,  // TODO
		TxNonce:  0,  // TODO
		Msg:      msg,
	}
	// XXX: ensure some canonical form
	b, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}
	return b
}

type SignBytesObject struct {
	ChainID  string `json:"chain_id"`
	AccNonce int64  `json:"acc_nonce"`
	TxNonce  int64  `json:"tx_nonce"`
	Msg      Msg    `json:"msg"`
}

//-------------------------------------

// Application function variable used to unmarshal transaction bytes
type TxDecoder func(txBytes []byte) (Tx, Error)
