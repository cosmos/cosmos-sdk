package baseapp

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SDK error codes
const (
	// ABCI error codes
	ABCICodeOK sdk.ABCICodeType = 0

	// Base error codes
	CodeOK                sdk.CodeType = 0
	CodeInternal          sdk.CodeType = 1
	CodeTxDecode          sdk.CodeType = 2
	CodeInvalidSequence   sdk.CodeType = 3
	CodeUnauthorized      sdk.CodeType = 4
	CodeInsufficientFunds sdk.CodeType = 5
	CodeUnknownRequest    sdk.CodeType = 6
	CodeInvalidAddress    sdk.CodeType = 7
	CodeInvalidPubKey     sdk.CodeType = 8
	CodeUnknownAddress    sdk.CodeType = 9
	CodeInsufficientCoins sdk.CodeType = 10
	CodeInvalidCoins      sdk.CodeType = 11

	CodespaceSDK sdk.CodespaceType = 2
)

// NOTE: Don't stringer this, we'll put better messages in later.
func CodeToDefaultMsg(code sdk.CodeType) string {
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
func ErrInternal(msg string) sdk.Error {
	return sdk.NewError(CodespaceSDK, CodeInternal, msg)
}
func ErrTxDecode(msg string) sdk.Error {
	return sdk.NewError(CodespaceSDK, CodeTxDecode, msg)
}
func ErrInvalidSequence(msg string) sdk.Error {
	return sdk.NewError(CodespaceSDK, CodeInvalidSequence, msg)
}
func ErrUnauthorized(msg string) sdk.Error {
	return sdk.NewError(CodespaceSDK, CodeUnauthorized, msg)
}
func ErrInsufficientFunds(msg string) sdk.Error {
	return sdk.NewError(CodespaceSDK, CodeInsufficientFunds, msg)
}
func ErrUnknownRequest(msg string) sdk.Error {
	return sdk.NewError(CodespaceSDK, CodeUnknownRequest, msg)
}
func ErrInvalidAddress(msg string) sdk.Error {
	return sdk.NewError(CodespaceSDK, CodeInvalidAddress, msg)
}
func ErrUnknownAddress(msg string) sdk.Error {
	return sdk.NewError(CodespaceSDK, CodeUnknownAddress, msg)
}
func ErrInvalidPubKey(msg string) sdk.Error {
	return sdk.NewError(CodespaceSDK, CodeInvalidPubKey, msg)
}
func ErrInsufficientCoins(msg string) sdk.Error {
	return sdk.NewError(CodespaceSDK, CodeInsufficientCoins, msg)
}
func ErrInvalidCoins(msg string) sdk.Error {
	return sdk.NewError(CodespaceSDK, CodeInvalidCoins, msg)
}
