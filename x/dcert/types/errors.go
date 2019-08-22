// nolint
package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type CodeType = sdk.CodeType

const (
	// default codespace for crisis module
	DefaultCodespace sdk.CodespaceType = ModuleName

	// CodeInvalidInput is the codetype for invalid input for the crisis module
	CodeInvalidInput CodeType = 12
)

// ErrNilSender -  no sender provided for the input
func ErrNilSender(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidInput, "sender address is nil")
}

// ErrNilSender -  no sender provided for the input
func ErrNilSuspend(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidInput, "address to suspend is nil")
}

// ErrNilSender -  no sender provided for the input
func ErrNilDescription(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidInput, "dcert description cannot be empty")
}
