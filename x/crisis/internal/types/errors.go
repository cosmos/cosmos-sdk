package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// default codespace for crisis module
	DefaultCodespace sdk.CodespaceType = ModuleName

	// CodeInvalidInput is the codetype for invalid input for the crisis module
	CodeInvalidInput sdk.CodeType = 103
)

// ErrNilSender -  no sender provided for the input
func ErrNilSender(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidInput, "sender address is nil")
}

// ErrUnknownInvariant -  unknown invariant provided
func ErrUnknownInvariant(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidInput, "unknown invariant")
}
