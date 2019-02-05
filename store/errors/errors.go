package errors

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	CodeOK             = sdk.CodeOK
	CodeInternal       = sdk.CodeInternal
	CodeTxDecode       = sdk.CodeTxDecode
	CodeUnknownRequest = sdk.CodeUnknownRequest

	CodespaceRoot = sdk.CodespaceRoot
)

type Error = sdk.Error

func ErrInternal(msg string) Error {
	return sdk.ErrInternal(msg)
}

func ErrTxDecode(msg string) Error {
	return sdk.ErrTxDecode(msg)
}

func ErrUnknownRequest(msg string) Error {
	return sdk.ErrUnknownRequest(msg)
}
