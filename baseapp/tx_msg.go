package baseapp

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
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
	ValidateBasic() sdk.Error

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
}

//__________________________________________________________

// TxDeocder unmarshals transaction bytes
type TxDecoder func(txBytes []byte) (Tx, sdk.Error)
