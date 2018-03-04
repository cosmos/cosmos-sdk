package types

import "encoding/json"

// Transactions messages must fulfill the Msg
type Msg interface {

	// Return the message type.
	// Must be alphanumeric or empty.
	Type() string

	// Get some property of the Msg.
	Get(key interface{}) (value interface{})

	// Get the canonical byte representation of the Msg.
	GetSignBytes() []byte

	// ValidateBasic does a simple validation check that
	// doesn't require access to any other information.
	ValidateBasic() Error

	// Signers returns the addrs of signers that must sign.
	// CONTRACT: All signatures must be present to be valid.
	// CONTRACT: Returns addrs in some deterministic order.
	GetSigners() []Address
}

//__________________________________________________________

// Transactions objects must fulfill the Tx
type Tx interface {

	// Gets the Msg.
	GetMsg() Msg

	// The address that pays the base fee for this message.  The fee is
	// deducted before the Msg is processed.
	GetFeePayer() Address

	// Signatures returns the signature of signers who signed the Msg.
	// CONTRACT: Length returned is same as length of
	// pubkeys returned from MsgKeySigners, and the order
	// matches.
	// CONTRACT: If the signature is missing (ie the Msg is
	// invalid), then the corresponding signature is
	// .Empty().
	GetSignatures() []StdSignature
}

var _ Tx = (*StdTx)(nil)

// StdTx is a standard way to wrap a Msg with Signatures.
// NOTE: the first signature is the FeePayer (Signatures must not be nil).
type StdTx struct {
	Msg        `json:"msg"`
	Signatures []StdSignature `json:"signatures"`
}

func NewStdTx(msg Msg, sigs []StdSignature) StdTx {
	return StdTx{
		Msg:        msg,
		Signatures: sigs,
	}
}

//nolint
func (tx StdTx) GetMsg() Msg                   { return tx.Msg }
func (tx StdTx) GetFeePayer() Address          { return tx.Signatures[0].PubKey.Address() } // XXX but PubKey is optional!
func (tx StdTx) GetSignatures() []StdSignature { return tx.Signatures }

// StdSignDoc is replay-prevention structure.
// It includes the result of msg.GetSignBytes(),
// as well as the ChainID (prevent cross chain replay)
// and the Sequence (prevent inchain replay).
type StdSignDoc struct {
	ChainID  string `json:"chain_id"`
	Sequence int64  `json:"sequence"`
	MsgBytes []byte `json:"msg_bytes"`
	AltBytes []byte `json:"alt_bytes"` // TODO: do we really want this ?
}

func StdSignBytes(chainID string, sequence int64, msg Msg) []byte {
	bz, err := json.Marshal(StdSignDoc{
		ChainID:  chainID,
		Sequence: sequence,
		MsgBytes: msg.GetSignBytes(),
	})
	if err != nil {
		panic(err)
	}
	return bz
}

//-------------------------------------

// Application function variable used to unmarshal transaction bytes
type TxDecoder func(txBytes []byte) (Tx, Error)
