package tx

import (
	"fmt"

	txsigning "cosmossdk.io/x/tx/signing"
	"cosmossdk.io/x/tx/signing/textual"

	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
)

// DefaultSignModes are the default sign modes enabled for protobuf transactions.
var DefaultSignModes = []signingtypes.SignMode{
	signingtypes.SignMode_SIGN_MODE_DIRECT,
	signingtypes.SignMode_SIGN_MODE_DIRECT_AUX,
	signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
	// We currently don't add SIGN_MODE_TEXTUAL as part of the default sign
	// modes, as it's not released yet (including the Ledger app). However,
	// textual's sign mode handler is already available in this package. If you
	// want to use textual for **TESTING** purposes, feel free to create a
	// handler that includes SIGN_MODE_TEXTUAL.
	// ref: Tracking issue for SIGN_MODE_TEXTUAL https://github.com/cosmos/cosmos-sdk/issues/11970
}

// makeSignModeHandler returns the default protobuf SignModeHandler supporting
// SIGN_MODE_DIRECT, SIGN_MODE_DIRECT_AUX and SIGN_MODE_LEGACY_AMINO_JSON.
func makeSignModeHandler(
	modes []signingtypes.SignMode,
	txt *textual.SignModeHandler,
	customSignModes ...txsigning.SignModeHandler,
) txsigning.SignModeHandler {
	if len(modes) < 1 {
		panic(fmt.Errorf("no sign modes enabled"))
	}

	handlers := make([]signing.SignModeHandler, len(modes)+len(customSignModes))

	// handle cosmos-sdk defined sign modes
	for i, mode := range modes {
		switch mode {
		case signingtypes.SignMode_SIGN_MODE_DIRECT:
			handlers[i] = signModeDirectHandler{}
		case signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON:
			handlers[i] = signModeLegacyAminoJSONHandler{}
		case signingtypes.SignMode_SIGN_MODE_TEXTUAL:
			handlers[i] = signModeTextualHandler{t: *txt}
		case signingtypes.SignMode_SIGN_MODE_DIRECT_AUX:
			handlers[i] = signModeDirectAuxHandler{}
		default:
			panic(fmt.Errorf("unsupported sign mode %+v", mode))
		}
	}

	// add custom sign modes
	for i, handler := range customSignModes {
		handlers[i+len(modes)] = handler
	}

	return signing.NewSignModeHandlerMap(
		modes[0],
		handlers,
	)
}
