//nolint
package coin

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/errors"
)

const (
	// Coin errors reserve 100 ~ 199.
	CodeInvalidInput   uint32 = 101
	CodeInvalidOutput  uint32 = 102
	CodeInvalidAddress uint32 = 103
	CodeUnknownAddress uint32 = 103
	CodeUnknownRequest uint32 = errors.CodeUnknownRequest
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
	case CodeUnknownRequest:
		return "Unknown request"
	default:
		return errors.CodeToDefaultLog(code)
	}
}

//----------------------------------------
// Error constructors

func ErrInvalidInput(log string) error {
	return newError(CodeInvalidInput, log)
}

func ErrInvalidOutput(log string) error {
	return newError(CodeInvalidOutput, log)
}

func ErrInvalidAddress(log string) error {
	return newError(CodeInvalidAddress, log)
}

func ErrUnknownAddress(log string) error {
	return newError(CodeUnknownAddress, log)
}

func ErrUnknownRequest(log string) error {
	return newError(CodeUnknownRequest, log)
}

//----------------------------------------
// Misc

func logOrDefault(log string, code uint32) string {
	if log != "" {
		return log
	} else {
		return codeToDefaultLog
	}
}

func newError(code uint32, log string) error {
	log = logOrDefaultLog(log, code)
	return errors.NewABCIError(code, log)
}
