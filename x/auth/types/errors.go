package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrorInvalidSigner        = sdkerrors.Register(ModuleName, 2, "tx intended signer does not match the given signer")
	ErrorInvalidGasAdjustment = sdkerrors.Register(ModuleName, 3, "invalid gas adjustment")
)
