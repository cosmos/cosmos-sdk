package tx

import (
	signing2 "github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/signing/direct"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// DefaultSignModeHandler returns the default protobuf SignModeHandler supporting
// SIGN_MODE_DIRECT and SIGN_MODE_LEGACY_AMINO_JSON.
func DefaultSignModeHandler() signing.SignModeHandler {
	return signing.NewSignModeHandlerMap(
		signing2.SignMode_SIGN_MODE_DIRECT,
		[]signing.SignModeHandler{
			authtypes.LegacyAminoJSONHandler{},
			direct.ModeHandler{},
		},
	)
}
