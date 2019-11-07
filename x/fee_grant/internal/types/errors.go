package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Codes for governance errors
const (
	DefaultCodespace = ModuleName
)

var (
	// ErrFeeLimitExceeded error if there are not enough allowance to cover the fees
	ErrFeeLimitExceeded = sdkerrors.Register(DefaultCodespace, 1, "fee limit exceeded")

	// ErrFeeLimitExpired error if the allowance has expired
	ErrFeeLimitExpired = sdkerrors.Register(DefaultCodespace, 2, "fee limit expired")

	// ErrInvalidDuration error if the Duration is invalid or doesn't match the expiration
	ErrInvalidDuration = sdkerrors.Register(DefaultCodespace, 3, "invalid duration")
)
