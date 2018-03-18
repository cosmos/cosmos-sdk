//nolint
package bank

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type CodeType = sdk.CodeType

const (
	// Coin errors reserve 100 - 199.
	CodeInvalidInput      CodeType = 100
	CodeInvalidOutput     CodeType = 101
	CodeInvalidAddress    CodeType = 102
	CodeUnknownAddress    CodeType = 103
	CodeInsufficientCoins CodeType = 104
	CodeInvalidCoins      CodeType = 105
	CodeUnknownRequest    CodeType = sdk.CodeUnknownRequest
)

// NOTE: Don't stringer this, we'll put better messages in later.
func codeToDefaultMsg(code CodeType) string {
	switch code {
	case CodeInvalidInput:
		return "Invalid input coins"
	case CodeInvalidOutput:
		return "Invalid output coins"
	case CodeInvalidAddress:
		return "Invalid address"
	case CodeUnknownAddress:
		return "Unknown address"
	case CodeInsufficientCoins:
		return "Insufficient coins"
	case CodeInvalidCoins:
		return "Invalid coins"
	case CodeUnknownRequest:
		return "Unknown request"
	default:
		return sdk.CodeToDefaultMsg(code)
	}
}

//----------------------------------------
// Error constructors

func ErrInvalidInput(msg string) sdk.Error {
	return newError(CodeInvalidInput, msg)
}

func ErrNoInputs() sdk.Error {
	return newError(CodeInvalidInput, "")
}

func ErrInvalidOutput(msg string) sdk.Error {
	return newError(CodeInvalidOutput, msg)
}

func ErrNoOutputs() sdk.Error {
	return newError(CodeInvalidOutput, "")
}

func ErrInvalidSequence(msg string) sdk.Error {
	return sdk.ErrInvalidSequence(msg)
}

func ErrInvalidAddress(msg string) sdk.Error {
	return newError(CodeInvalidAddress, msg)
}

func ErrUnknownAddress(msg string) sdk.Error {
	return newError(CodeUnknownAddress, msg)
}

func ErrInsufficientCoins(msg string) sdk.Error {
	return newError(CodeInsufficientCoins, msg)
}

func ErrInvalidCoins(msg string) sdk.Error {
	return newError(CodeInvalidCoins, msg)
}

func ErrUnknownRequest(msg string) sdk.Error {
	return newError(CodeUnknownRequest, msg)
}

//----------------------------------------

func msgOrDefaultMsg(msg string, code CodeType) string {
	if msg != "" {
		return msg
	} else {
		return codeToDefaultMsg(code)
	}
}

func newError(code CodeType, msg string) sdk.Error {
	msg = msgOrDefaultMsg(msg, code)
	return sdk.NewError(code, msg)
}
