package simplestake

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// simple stake errors reserve 300 ~ 399.
const (
	DefaultCodespace sdk.CodespaceType = 4

	// simplestake errors reserve 300 - 399.
	CodeEmptyValidator        sdk.CodeType = 300
	CodeInvalidUnbond         sdk.CodeType = 301
	CodeEmptyStake            sdk.CodeType = 302
	CodeIncorrectStakingToken sdk.CodeType = 303
)

// nolint
func ErrIncorrectStakingToken(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeIncorrectStakingToken, "")
}
func ErrEmptyValidator(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeEmptyValidator, "")
}
func ErrInvalidUnbond(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeInvalidUnbond, "")
}
func ErrEmptyStake(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeEmptyStake, "")
}

// -----------------------------
// Helpers

// nolint: unparam
func newError(codespace sdk.CodespaceType, code sdk.CodeType, msg string) sdk.Error {
	return sdk.NewError(codespace, code, msg)
}
