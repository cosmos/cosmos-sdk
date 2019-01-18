package oracle

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Oracle errors reserve 1101-1199
const (
	CodeNotValidator     sdk.CodeType = 1101
	CodeAlreadyProcessed sdk.CodeType = 1102
	CodeAlreadySigned    sdk.CodeType = 1103
	CodeUnknownRequest   sdk.CodeType = sdk.CodeUnknownRequest
)

// ----------------------------------------
// Error constructors

// ErrNotValidator called when the signer of a Msg is not a validator
func ErrNotValidator(codespace sdk.CodespaceType, address sdk.AccAddress) sdk.Error {
	return sdk.NewError(codespace, CodeNotValidator, address.String())
}

// ErrAlreadyProcessed called when a payload is already processed
func ErrAlreadyProcessed(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeAlreadyProcessed, "")
}

// ErrAlreadySigned called when the signer is trying to double signing
func ErrAlreadySigned(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeAlreadySigned, "")
}
