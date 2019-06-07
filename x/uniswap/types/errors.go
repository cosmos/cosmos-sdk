package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	DefaultCodespace sdk.CodespaceType = ModuleName

	CodeNoDenom               sdk.CodeType = 101
	CodeExchangeAlreadyExists sdk.CodeType = 102
	CodeEqualDenom            sdk.CodeType = 103
	CodeInvalidBound          sdk.CodeType = 104
	CodeInvalidDeadline       sdk.CodeType = 105
	CodeInsufficientAmount    sdk.CodeType = 106
)

func ErrNoDenom(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeNoDenom, "denomination is empty")
}

func ErrExchangeAlreadyExists(codespace sdk.CodespaceType, msg string) sdk.Error {
	if msg != "" {
		return sdk.NewError(codespace, CodeExchangeAlreadyExists, msg)
	}
	return sdk.NewError(codespace, CodeExchangeAlreadyExists, "exchange already exists")
}

func ErrEqualDenom(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeEqualDenom, "coin and swap denomination are equal")
}

func ErrInvalidBound(codespace sdk.CodespaceType, msg string) sdk.Error {
	if msg != "" {
		return sdk.NewError(codespace, CodeInvalidBound, msg)
	}
	return sdk.NewError(codespace, CodeInvalidBound, "bound is not positive")
}

func ErrInvalidDeadline(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidDeadline, "deadline not initialized")
}

func ErrInsufficientAmount(codespace sdk.CodespaceType, msg string) sdk.Error {
	if msg != "" {
		return sdk.NewError(codespace, CodeInsufficientAmount, msg)
	}
	return sdk.NewError(codespace, CodeInsufficientAmount, "insufficient amount provided")
}
