//nolint
package bank

import (
	bapp "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	DefaultCodespace bapp.CodespaceType = 2

	CodeInvalidInput  bapp.CodeType = 101
	CodeInvalidOutput bapp.CodeType = 102
)

// NOTE: Don't stringer this, we'll put better messages in later.
func codeToDefaultMsg(code bapp.CodeType) string {
	switch code {
	case CodeInvalidInput:
		return "Invalid input coins"
	case CodeInvalidOutput:
		return "Invalid output coins"
	default:
		return sdk.CodeToDefaultMsg(code)
	}
}

//----------------------------------------
// Error constructors

func ErrInvalidInput(codespace bapp.CodespaceType, msg string) bapp.Error {
	return newError(codespace, CodeInvalidInput, msg)
}

func ErrNoInputs(codespace bapp.CodespaceType) bapp.Error {
	return newError(codespace, CodeInvalidInput, "")
}

func ErrInvalidOutput(codespace bapp.CodespaceType, msg string) bapp.Error {
	return newError(codespace, CodeInvalidOutput, msg)
}

func ErrNoOutputs(codespace bapp.CodespaceType) bapp.Error {
	return newError(codespace, CodeInvalidOutput, "")
}

//----------------------------------------

func msgOrDefaultMsg(msg string, code bapp.CodeType) string {
	if msg != "" {
		return msg
	}
	return codeToDefaultMsg(code)
}

func newError(codespace bapp.CodespaceType, code bapp.CodeType, msg string) bapp.Error {
	msg = msgOrDefaultMsg(msg, code)
	return bapp.NewError(codespace, code, msg)
}
