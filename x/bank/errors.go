package bank

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CodeType definition
type CodeType = sdk.CodeType

// Bank errors reserve 100 ~ 199.
const (
	DefaultCodespace sdk.CodespaceType = ModuleName

	CodeSendDisabled         CodeType = 101
	CodeInvalidInputsOutputs CodeType = 102
	CodeUnknownTokenHolder   CodeType = 103
)

// ErrNoInputs is an error
func ErrNoInputs(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidInputsOutputs, "no inputs to send transaction")
}

// ErrNoOutputs is an error
func ErrNoOutputs(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidInputsOutputs, "no outputs to send transaction")
}

// ErrInputOutputMismatch is an error
func ErrInputOutputMismatch(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidInputsOutputs, "sum inputs != sum outputs")
}

// ErrSendDisabled is an error
func ErrSendDisabled(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeSendDisabled, "send transactions are currently disabled")
}

// ErrUnknownTokenHolder is an error
func ErrUnknownTokenHolder(codespace sdk.CodespaceType, msg string) sdk.Error {
	if msg != "" {
		return sdk.NewError(codespace, CodeUnknownTokenHolder, msg)
	}
	return sdk.NewError(codespace, CodeUnknownTokenHolder, "unknown token holder")
}
