package types

import (
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
)

// Wrapper around baseapp.Error
type Error = struct {
	*baseapp.SdkError
}

// // NOTE: Don't stringer this, we'll put better messages in later.
// func CodeToDefaultMsg(code baseapp.CodeType) string {
// 	switch code {
// 	case CodeInternal:
// 		return "Internal error"
// 	case CodeTxDecode:
// 		return "Tx parse error"
// 	case CodeInvalidSequence:
// 		return "Invalid sequence"
// 	case CodeUnauthorized:
// 		return "Unauthorized"
// 	case CodeInsufficientFunds:
// 		return "Insufficent funds"
// 	case CodeUnknownRequest:
// 		return "Unknown request"
// 	case CodeInvalidAddress:
// 		return "Invalid address"
// 	case CodeInvalidPubKey:
// 		return "Invalid pubkey"
// 	case CodeUnknownAddress:
// 		return "Unknown address"
// 	case CodeInsufficientCoins:
// 		return "Insufficient coins"
// 	case CodeInvalidCoins:
// 		return "Invalid coins"
// 	default:
// 		return fmt.Sprintf("Unknown code %d", code)
// 	}
// }

//--------------------------------------------------------------------------------
// All errors are created via constructors so as to enable us to hijack them
// and inject stack traces if we really want to.

// nolint
func ErrInternal(msg string) Error {
	return newErrorWithRootCodespace(CodeInternal, msg)
}
func ErrTxDecode(msg string) Error {
	return newErrorWithRootCodespace(CodeTxDecode, msg)
}
func ErrInvalidSequence(msg string) Error {
	return newErrorWithRootCodespace(CodeInvalidSequence, msg)
}
func ErrUnauthorized(msg string) Error {
	return newErrorWithRootCodespace(CodeUnauthorized, msg)
}
func ErrInsufficientFunds(msg string) Error {
	return newErrorWithRootCodespace(CodeInsufficientFunds, msg)
}
func ErrUnknownRequest(msg string) Error {
	return newErrorWithRootCodespace(CodeUnknownRequest, msg)
}
func ErrInvalidAddress(msg string) Error {
	return newErrorWithRootCodespace(CodeInvalidAddress, msg)
}
func ErrUnknownAddress(msg string) Error {
	return newErrorWithRootCodespace(CodeUnknownAddress, msg)
}
func ErrInvalidPubKey(msg string) Error {
	return newErrorWithRootCodespace(CodeInvalidPubKey, msg)
}
func ErrInsufficientCoins(msg string) Error {
	return newErrorWithRootCodespace(CodeInsufficientCoins, msg)
}
func ErrInvalidCoins(msg string) Error {
	return newErrorWithRootCodespace(CodeInvalidCoins, msg)
}
