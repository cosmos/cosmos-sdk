package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/baseapp"
)

// SDK error codes
const (
	// ABCI error codes
	ABCICodeOK baseapp.ABCICodeType = 0

	// Base error codes
	CodeOK                baseapp.CodeType = 0
	CodeInternal          baseapp.CodeType = 1
	CodeTxDecode          baseapp.CodeType = 2
	CodeInvalidSequence   baseapp.CodeType = 3
	CodeUnauthorized      baseapp.CodeType = 4
	CodeInsufficientFunds baseapp.CodeType = 5
	CodeUnknownRequest    baseapp.CodeType = 6
	CodeInvalidAddress    baseapp.CodeType = 7
	CodeInvalidPubKey     baseapp.CodeType = 8
	CodeUnknownAddress    baseapp.CodeType = 9
	CodeInsufficientCoins baseapp.CodeType = 10
	CodeInvalidCoins      baseapp.CodeType = 11

	CodespaceSDK baseapp.CodespaceType = 2
)

// NOTE: Don't stringer this, we'll put better messages in later.
func CodeToDefaultMsg(code baseapp.CodeType) string {
	switch code {
	case CodeInternal:
		return "Internal error"
	case CodeTxDecode:
		return "Tx parse error"
	case CodeInvalidSequence:
		return "Invalid sequence"
	case CodeUnauthorized:
		return "Unauthorized"
	case CodeInsufficientFunds:
		return "Insufficent funds"
	case CodeUnknownRequest:
		return "Unknown request"
	case CodeInvalidAddress:
		return "Invalid address"
	case CodeInvalidPubKey:
		return "Invalid pubkey"
	case CodeUnknownAddress:
		return "Unknown address"
	case CodeInsufficientCoins:
		return "Insufficient coins"
	case CodeInvalidCoins:
		return "Invalid coins"
	default:
		return fmt.Sprintf("Unknown code %d", code)
	}
}

//--------------------------------------------------------------------------------
// All errors are created via constructors so as to enable us to hijack them
// and inject stack traces if we really want to.

// nolint
func ErrInternal(msg string) baseapp.Error {
	return baseapp.NewError(CodespaceSDK, CodeInternal, msg)
}
func ErrTxDecode(msg string) baseapp.Error {
	return baseapp.NewError(CodespaceSDK, CodeTxDecode, msg)
}
func ErrInvalidSequence(msg string) baseapp.Error {
	return baseapp.NewError(CodespaceSDK, CodeInvalidSequence, msg)
}
func ErrUnauthorized(msg string) baseapp.Error {
	return baseapp.NewError(CodespaceSDK, CodeUnauthorized, msg)
}
func ErrInsufficientFunds(msg string) baseapp.Error {
	return baseapp.NewError(CodespaceSDK, CodeInsufficientFunds, msg)
}
func ErrUnknownRequest(msg string) baseapp.Error {
	return baseapp.NewError(CodespaceSDK, CodeUnknownRequest, msg)
}
func ErrInvalidAddress(msg string) baseapp.Error {
	return baseapp.NewError(CodespaceSDK, CodeInvalidAddress, msg)
}
func ErrUnknownAddress(msg string) baseapp.Error {
	return baseapp.NewError(CodespaceSDK, CodeUnknownAddress, msg)
}
func ErrInvalidPubKey(msg string) baseapp.Error {
	return baseapp.NewError(CodespaceSDK, CodeInvalidPubKey, msg)
}
func ErrInsufficientCoins(msg string) baseapp.Error {
	return baseapp.NewError(CodespaceSDK, CodeInsufficientCoins, msg)
}
func ErrInvalidCoins(msg string) baseapp.Error {
	return baseapp.NewError(CodespaceSDK, CodeInvalidCoins, msg)
}
