package types

import (
	"encoding/json"
)

// Transactions messages must fulfill the Msg
type Msg interface {

	// Return the message type.
	// Must be alphanumeric or empty.
	Type() string

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
	GetMsgs() []Msg

	// Gets the memo.
	GetMemo() string
}

//__________________________________________________________

// TxDecoder unmarshals transaction bytes
type TxDecoder func(txBytes []byte) (Tx, Error)

//__________________________________________________________

var _ Msg = (*TestMsg)(nil)

// msg type for testing
type TestMsg struct {
	signers []Address
}

func NewTestMsg(addrs ...Address) *TestMsg {
	return &TestMsg{
		signers: addrs,
	}
}

//nolint
func (msg *TestMsg) Type() string { return "TestMsg" }
func (msg *TestMsg) GetSignBytes() []byte {
	bz, err := json.Marshal(msg.signers)
	if err != nil {
		panic(err)
	}
	return bz
}
func (msg *TestMsg) ValidateBasic() Error { return nil }
func (msg *TestMsg) GetSigners() []Address {
	return msg.signers
}
