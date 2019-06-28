package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	DefaultCodespace sdk.CodespaceType = ModuleName

	CodeReservePoolAlreadyExists sdk.CodeType = 101
	CodeEqualDenom               sdk.CodeType = 102
	CodeInvalidDeadline          sdk.CodeType = 103
	CodeNotPositive              sdk.CodeType = 104
	CodeConstraintNotMet         sdk.CodeType = 105
)

func ErrReservePoolAlreadyExists(codespace sdk.CodespaceType, msg string) sdk.Error {
	if msg != "" {
		return sdk.NewError(codespace, CodeReservePoolAlreadyExists, msg)
	}
	return sdk.NewError(codespace, CodeReservePoolAlreadyExists, "reserve pool already exists")
}

func ErrEqualDenom(codespace sdk.CodespaceType, msg string) sdk.Error {
	if msg != "" {
		sdk.NewError(codespace, CodeEqualDenom, msg)
	}
	return sdk.NewError(codespace, CodeEqualDenom, "input and output denomination are equal")
}

func ErrInvalidDeadline(codespace sdk.CodespaceType, msg string) sdk.Error {
	if msg != "" {
		return sdk.NewError(codespace, CodeInvalidDeadline, msg)
	}
	return sdk.NewError(codespace, CodeInvalidDeadline, "invalid deadline")
}

func ErrNotPositive(codespace sdk.CodespaceType, msg string) sdk.Error {
	if msg != "" {
		return sdk.NewError(codespace, CodeNotPositive, msg)
	}
	return sdk.NewError(codespace, CodeNotPositive, "amount is not positive")
}

func ErrConstraintNotMet(codespace sdk.CodespaceType, msg string) sdk.Error {
	if msg != "" {
		return sdk.NewError(codespace, CodeConstraintNotMet, msg)
	}
	return sdk.NewError(codespace, CodeConstraintNotMet, "constraint not met")
}
