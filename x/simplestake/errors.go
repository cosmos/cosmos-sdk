package staking

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// Staking errors reserve 300 - 399.
	CodeEmptyValidator        sdk.CodeType = 300
	CodeInvalidUnbond         sdk.CodeType = 301
	CodeEmptyStake            sdk.CodeType = 302
	CodeIncorrectStakingToken sdk.CodeType = 303
)

func ErrIncorrectStakingToken() sdk.Error {
	return newError(CodeIncorrectStakingToken, "")
}

func ErrEmptyValidator() sdk.Error {
	return newError(CodeEmptyValidator, "")
}

func ErrInvalidUnbond() sdk.Error {
	return newError(CodeInvalidUnbond, "")
}

func ErrEmptyStake() sdk.Error {
	return newError(CodeEmptyStake, "")
}

// -----------------------------
// Helpers

func newError(code sdk.CodeType, msg string) sdk.Error {
	return sdk.NewError(code, msg)
}
