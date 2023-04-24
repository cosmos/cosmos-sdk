package tx

import (
	txsigning "cosmossdk.io/x/tx/signing"
	"cosmossdk.io/x/tx/signing/aminojson"
	"cosmossdk.io/x/tx/signing/direct"
	"cosmossdk.io/x/tx/signing/directaux"
	"cosmossdk.io/x/tx/signing/textual"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
)

type SignModeOptions struct {
	// Textual are options for SIGN_MODE_TEXTUAL
	Textual *textual.SignModeOptions
	// DirectAux are options for SIGN_MODE_DIRECT_AUX
	DirectAux *directaux.SignModeHandlerOptions
	// AminoJSON are options for SIGN_MODE_LEGACY_AMINO_JSON
	AminoJSON *aminojson.SignModeHandlerOptions
	// Direct is the SignModeHandler for SIGN_MODE_DIRECT since it takes options
	Direct *direct.SignModeHandler
}

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
	opts SignModeOptions,
	customSignModes ...txsigning.SignModeHandler,
) *txsigning.HandlerMap {
	var handlers []txsigning.SignModeHandler
	if opts.Direct != nil {
		handlers = append(handlers, opts.Direct)
	}
	if opts.Textual != nil {
		h, err := textual.NewSignModeHandler(*opts.Textual)
		if err != nil {
			panic(err)
		}
		handlers = append(handlers, h)
	}
	if opts.DirectAux != nil {
		h, err := directaux.NewSignModeHandler(*opts.DirectAux)
		if err != nil {
			panic(err)
		}
		handlers = append(handlers, h)
	}
	if opts.AminoJSON != nil {
		handlers = append(handlers, aminojson.NewSignModeHandler(*opts.AminoJSON))
	}
	handlers = append(handlers, customSignModes...)
	return txsigning.NewHandlerMap(handlers...)
}
