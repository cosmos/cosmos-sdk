package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type CodeType = sdk.CodeType

const (
	DefaultCodespace sdk.CodespaceType = 11

	// simple errors reserve 1101 ~ 1199.
	CodeInvalidGenesis CodeType = 1101
)

func codeToDefaultMsg(code CodeType) string {
	switch code {
	case CodeInvalidGenesis:
		return "Invalid genesis"
	default:
		return sdk.CodeToDefaultMsg(code)
	}
}

//----------------------------------------
// Error constructors

// nolint
func ErrInvalidName(msg string) sdk.Error {
	return newError(DefaultCodespace, CodeInvalidGenesis, msg)
}

//----------------------------------------

func msgOrDefaultMsg(msg string, code CodeType) string {
	if msg != "" {
		return msg
	}
	return codeToDefaultMsg(code)
}

func newError(codespace sdk.CodespaceType, code CodeType, msg string) sdk.Error {
	msg = msgOrDefaultMsg(msg, code)
	return sdk.NewError(codespace, code, msg)
}
