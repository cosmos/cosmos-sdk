package feegrant

import (
	errorsmod "cosmossdk.io/errors"
)

// Codes for governance errors
const (
	DefaultCodespace = ModuleName
)

var (
	// ErrFeeLimitExceeded error if there are not enough allowance to cover the fees
	ErrFeeLimitExceeded = errorsmod.Register(DefaultCodespace, 2, "fee limit exceeded")
	// ErrFeeLimitExpired error if the allowance has expired
	ErrFeeLimitExpired = errorsmod.Register(DefaultCodespace, 3, "fee allowance expired")
	// ErrInvalidDuration error if the Duration is invalid or doesn't match the expiration
	ErrInvalidDuration = errorsmod.Register(DefaultCodespace, 4, "invalid duration")
	// ErrNoAllowance error if there is no allowance for that pair
	ErrNoAllowance = errorsmod.Register(DefaultCodespace, 5, "no allowance")
	// ErrNoMessages error if there is no message
	ErrNoMessages = errorsmod.Register(DefaultCodespace, 6, "allowed messages are empty")
	// ErrMessageNotAllowed error if message is not allowed
	ErrMessageNotAllowed = errorsmod.Register(DefaultCodespace, 7, "message not allowed")
)
