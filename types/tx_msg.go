package types

import (
	"encoding/json"
)

// Transactions messages must fulfill the Msg
type Msg interface {

	// Return the message type.
	// Must be alphanumeric or empty.
	Type() string

	// Returns a human-readable string for the message, intended for utilization
	// within tags
	Name() string

	// ValidateBasic does a simple validation check that
	// doesn't require access to any other information.
	ValidateBasic() Error

	// Get the canonical byte representation of the Msg.
	GetSignBytes() []byte

	// Signers returns the addrs of signers that must sign.
	// CONTRACT: All signatures must be present to be valid.
	// CONTRACT: Returns addrs in some deterministic order.
	GetSigners() []AccAddress
}

//__________________________________________________________

// Transactions objects must fulfill the Tx
type Tx interface {

	// Gets the Msg.
	GetMsgs() []Msg
}

//__________________________________________________________

// TxDecoder unmarshals transaction bytes
type TxDecoder func(txBytes []byte) (Tx, Error)

//__________________________________________________________

var _ Msg = (*TestMsg)(nil)

// msg type for testing
type TestMsg struct {
	signers []AccAddress
}

func NewTestMsg(addrs ...AccAddress) *TestMsg {
	return &TestMsg{
		signers: addrs,
	}
}

//nolint
func (msg *TestMsg) Type() string { return "TestMsg" }
func (msg *TestMsg) Name() string { return "Test message" }
func (msg *TestMsg) GetSignBytes() []byte {
	bz, err := json.Marshal(msg.signers)
	if err != nil {
		panic(err)
	}
	return MustSortJSON(bz)
}
func (msg *TestMsg) ValidateBasic() Error { return nil }
func (msg *TestMsg) GetSigners() []AccAddress {
	return msg.signers
}
