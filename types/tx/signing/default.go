package signing

import (
	types "github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func DefaultSignModeHandler() types.SignModeHandler {
	return NewHandlerMap(types.SignMode_SIGN_MODE_DIRECT, []types.SignModeHandler{
		DirectModeHandler{},
		authtypes.LegacyAminoJSONHandler{},
	})
}
