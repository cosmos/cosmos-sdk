package signing

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
)

// SignModeHandler defines a interface to be implemented by types which will handle
// SignMode's by generating sign bytes from a Tx and SignerData
type SignModeHandler interface {
	// DefaultMode is the default mode that is to be used with this handler if no
	// other mode is specified. This can be useful for testing and CLI usage
	DefaultMode() txtypes.SignMode

	// Modes is the list of modes supporting by this handler
	Modes() []txtypes.SignMode

	// GetSignBytes returns the sign bytes for the provided SignMode, SignerData and Tx,
	// or an error
	GetSignBytes(mode txtypes.SignMode, data SignerData, tx sdk.Tx) ([]byte, error)
}

// SignerData is the specific information needed to sign a transaction that generally
// isn't included in the transaction body itself
type SignerData struct {
	// ChainID is the chain that this transaction is targeted
	ChainID string

	// AccountNumber is the account number of the signer
	AccountNumber uint64

	// AccountSequence is the account sequence number of the signer that is used
	// for replay protection
	AccountSequence uint64
}
