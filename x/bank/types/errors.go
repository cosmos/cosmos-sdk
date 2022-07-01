package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/bank module sentinel errors
var (
	// Error codes 2-4 were already assigned and removed, but should not be recycled.
	ErrSendDisabled          = sdkerrors.Register(ModuleName, 5, "send transactions are disabled")
	ErrDenomMetadataNotFound = sdkerrors.Register(ModuleName, 6, "client denom metadata not found")
	ErrInvalidKey            = sdkerrors.Register(ModuleName, 7, "invalid key")
)
