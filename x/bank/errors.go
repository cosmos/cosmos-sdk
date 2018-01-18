//nolint
package bank

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// Coin errors reserve 100 ~ 199.
	CodeInvalidInput      uint32 = 101
	CodeInvalidOutput     uint32 = 102
	CodeInvalidAddress    uint32 = 103
	CodeUnknownAddress    uint32 = 104
	CodeInsufficientCoins uint32 = 105
	CodeInvalidCoins      uint32 = 106
	CodeInvalidSequence   uint32 = 107
	CodeUnknownRequest    uint32 = sdk.CodeUnknownRequest
)

// NOTE: Don't stringer this, we'll put better messages in later.
func codeToDefaultLog(code uint32) string {
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
	case CodeInvalidSequence:
		return "Invalid sequence"
	case CodeUnknownRequest:
		return "Unknown request"
	default:
		return sdk.CodeToDefaultLog(code)
	}
}

//----------------------------------------
// Error constructors

func ErrInvalidInput(log string) sdk.Error {
	return newError(CodeInvalidInput, log)
}

func ErrNoInputs() sdk.Error {
	return newError(CodeInvalidInput, "")
}

func ErrInvalidOutput(log string) sdk.Error {
	return newError(CodeInvalidOutput, log)
}

func ErrNoOutputs() sdk.Error {
	return newError(CodeInvalidOutput, "")
}

func ErrInvalidSequence(seq int64) sdk.Error {
	return newError(CodeInvalidSequence, "")
}

func ErrInvalidAddress(log string) sdk.Error {
	return newError(CodeInvalidAddress, log)
}

func ErrUnknownAddress(log string) sdk.Error {
	return newError(CodeUnknownAddress, log)
}

func ErrInsufficientCoins(log string) sdk.Error {
	return newError(CodeInsufficientCoins, log)
}

func ErrInvalidCoins(log string) sdk.Error {
	return newError(CodeInvalidCoins, log)
}

func ErrUnknownRequest(log string) sdk.Error {
	return newError(CodeUnknownRequest, log)
}

//----------------------------------------

func logOrDefaultLog(log string, code uint32) string {
	if log != "" {
		return log
	} else {
		return codeToDefaultLog(code)
	}
}

func newError(code uint32, log string) sdk.Error {
	log = logOrDefaultLog(log, code)
	return sdk.NewError(code, log)
}
