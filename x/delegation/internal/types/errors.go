package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Codes for governance errors
const (
	DefaultCodespace sdk.CodespaceType = ModuleName

	CodeFeeLimitExceeded sdk.CodeType = 1
	CodeFeeLimitExpired  sdk.CodeType = 2
	CodeInvalidPeriod    sdk.CodeType = 3
	CodeNonPositiveCoins sdk.CodeType = 4
)

// ErrFeeLimitExceeded error if there are not enough allowance to cover the fees
func ErrFeeLimitExceeded() sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeFeeLimitExceeded, "fee limit exceeded")
}

// ErrFeeLimitExpired error if the allowance has expired
func ErrFeeLimitExpired() sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeFeeLimitExpired, "fee limit expired")
}

// ErrInvalidPeriod error if the period is invalid or doesn't match the expiration
func ErrInvalidPeriod(reason string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeInvalidPeriod, reason)
}

// ErrNonPositiveCoins error if some fees or allowance are non positive
func ErrNonPositiveCoins() sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeNonPositiveCoins, "non positive coin amount")
}
