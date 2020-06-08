package types

import (
	"encoding/json"

	"github.com/gogo/protobuf/proto"

	"github.com/tendermint/tendermint/crypto"
)

type (
	// Msg defines the interface a transaction message must fulfill.
	Msg interface {
		proto.Message

		// Return the message type.
		// Must be alphanumeric or empty.
		Route() string

		// Returns a human-readable string for the message, intended for utilization
		// within tags
		Type() string

		// ValidateBasic does a simple validation check that
		// doesn't require access to any other information.
		ValidateBasic() error

		// Get the canonical byte representation of the Msg.
		GetSignBytes() []byte

		// Signers returns the addrs of signers that must sign.
		// CONTRACT: All signatures must be present to be valid.
		// CONTRACT: Returns addrs in some deterministic order.
		GetSigners() []AccAddress
	}

	// Fee defines an interface for an application application-defined concrete
	// transaction type to be able to set and return the transaction fee.
	Fee interface {
		GetGas() uint64
		GetAmount() Coins
	}

	// Signature defines an interface for an application application-defined
	// concrete transaction type to be able to set and return transaction signatures.
	Signature interface {
		GetPubKey() crypto.PubKey
		GetSignature() []byte
	}

	// Tx defines the interface a transaction must fulfill.
	Tx interface {
		// Gets the all the transaction's messages.
		GetMsgs() []Msg

		// ValidateBasic does a simple and lightweight validation check that doesn't
		// require access to any other information.
		ValidateBasic() error
	}
)

// TxDecoder unmarshals transaction bytes
type TxDecoder func(txBytes []byte) (Tx, error)

// TxEncoder marshals transaction to bytes
type TxEncoder func(tx Tx) ([]byte, error)

//__________________________________________________________

var _ Msg = (*TestMsg)(nil)

// msg type for testing
type TestMsg struct {
	signers []AccAddress
}

// dummy implementation of proto.Message
func (msg *TestMsg) Reset()         {}
func (msg *TestMsg) String() string { return "TODO" }
func (msg *TestMsg) ProtoMessage()  {}

func NewTestMsg(addrs ...AccAddress) *TestMsg {
	return &TestMsg{
		signers: addrs,
	}
}

func (msg *TestMsg) Route() string { return "TestMsg" }
func (msg *TestMsg) Type() string  { return "Test message" }
func (msg *TestMsg) GetSignBytes() []byte {
	bz, err := json.Marshal(msg.signers)
	if err != nil {
		panic(err)
	}
	return MustSortJSON(bz)
}
func (msg *TestMsg) ValidateBasic() error { return nil }
func (msg *TestMsg) GetSigners() []AccAddress {
	return msg.signers
}
