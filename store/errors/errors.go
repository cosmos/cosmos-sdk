package errors

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// nolint - reexport
const (
	CodeOK             = sdk.CodeOK
	CodeInternal       = sdk.CodeInternal
	CodeTxDecode       = sdk.CodeTxDecode
	CodeUnknownRequest = sdk.CodeUnknownRequest

	CodespaceRoot = sdk.CodespaceRoot
)

// nolint - reexport
type Error = sdk.Error

// nolint - reexport
func ErrInternal(msg string) Error {
	return sdk.ErrInternal(msg)
}
func ErrTxDecode(msg string) Error {
	return sdk.ErrTxDecode(msg)
}
func ErrUnknownRequest(msg string) Error {
	return sdk.ErrUnknownRequest(msg)
}
