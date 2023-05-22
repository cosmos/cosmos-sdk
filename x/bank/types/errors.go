package types

import (
	errorsmod "cosmossdk.io/errors"
)

// x/bank module sentinel errors
var (
	ErrNoInputs              = errorsmod.Register(ModuleName, 2, "no inputs to send transaction")
	ErrNoOutputs             = errorsmod.Register(ModuleName, 3, "no outputs to send transaction")
	ErrInputOutputMismatch   = errorsmod.Register(ModuleName, 4, "sum inputs != sum outputs")
	ErrSendDisabled          = errorsmod.Register(ModuleName, 5, "send transactions are disabled")
	ErrDenomMetadataNotFound = errorsmod.Register(ModuleName, 6, "client denom metadata not found")
	ErrInvalidKey            = errorsmod.Register(ModuleName, 7, "invalid key")
	ErrDuplicateEntry        = errorsmod.Register(ModuleName, 8, "duplicate entry")
	ErrMultipleSenders       = errorsmod.Register(ModuleName, 9, "multiple senders not allowed")
)
