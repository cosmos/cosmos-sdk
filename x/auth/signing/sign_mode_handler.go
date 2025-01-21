package signing

import (
	"cosmossdk.io/x/tx/signing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdksigning "github.com/cosmos/cosmos-sdk/types/tx/signing"
)

// SignModeHandler defines a interface to be implemented by types which will handle
// SignMode's by generating sign bytes from a Tx and SignerData
type SignModeHandler interface {
	// DefaultMode is the default mode that is to be used with this handler if no
	// other mode is specified. This can be useful for testing and CLI usage
	DefaultMode() sdksigning.SignMode

	// Modes is the list of modes supporting by this handler
	Modes() []sdksigning.SignMode

	// GetSignBytes returns the sign bytes for the provided SignMode, SignerData and Tx,
	// or an error
	GetSignBytes(mode sdksigning.SignMode, data signing.SignerData, tx sdk.Tx) ([]byte, error)
}
