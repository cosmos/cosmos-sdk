package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Codes for governance errors
const (
	DefaultCodespace sdk.CodespaceType = ModuleName

	CodeFeeLimitExceeded sdk.CodeType = 1
	CodeFeeLimitExpired  sdk.CodeType = 2
	CodeInvalidDuration  sdk.CodeType = 3
)

// ErrFeeLimitExceeded error if there are not enough allowance to cover the fees
func ErrFeeLimitExceeded() sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeFeeLimitExceeded, "fee limit exceeded")
}

// ErrFeeLimitExpired error if the allowance has expired
func ErrFeeLimitExpired() sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeFeeLimitExpired, "fee limit expired")
}

// ErrInvalidDuration error if the Duration is invalid or doesn't match the expiration
func ErrInvalidDuration(reason string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeInvalidDuration, reason)
}
