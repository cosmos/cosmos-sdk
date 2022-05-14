package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/bank module sentinel errors
var (
	ErrNoInputs              = sdkerrors.Register(ModuleName, 2, "no inputs to send transaction")
	ErrNoOutputs             = sdkerrors.Register(ModuleName, 3, "no outputs to send transaction")
	ErrInputOutputMismatch   = sdkerrors.Register(ModuleName, 4, "sum inputs != sum outputs")
	ErrSendDisabled          = sdkerrors.Register(ModuleName, 5, "send transactions are disabled")
	ErrDenomMetadataNotFound = sdkerrors.Register(ModuleName, 6, "client denom metadata not found")
	ErrInvalidKey            = sdkerrors.Register(ModuleName, 7, "invalid key")

	// Periodic auth errors
	// ErrSpendLimitExceeded error if there are not enough allowance to cover the fees
	ErrSpendLimitExceeded = sdkerrors.Register(ModuleName, 8, "spend limit exceeded")
	// ErrSpendLimitExpired error if the allowance has expired
	ErrSpendLimitExpired = sdkerrors.Register(ModuleName, 9, "spend allowance expired")
	// ErrInvalidDuration error if the Duration is invalid or doesn't match the expiration
	ErrInvalidDuration = sdkerrors.Register(ModuleName, 10, "invalid duration")
	// ErrNoAllowance error if there is no allowance for that pair
	ErrNoAllowance = sdkerrors.Register(ModuleName, 11, "no allowance")
	// ErrNoMessages error if there is no message
	ErrNoMessages = sdkerrors.Register(ModuleName, 12, "allowed messages are empty")
	// ErrMessageNotAllowed error if message is not allowed
	ErrMessageNotAllowed = sdkerrors.Register(ModuleName, 13, "message not allowed")
)
