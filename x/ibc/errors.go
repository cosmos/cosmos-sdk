package ibc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// IBC errors reserve 200 - 299.
	CodeInvalidSequence sdk.CodeType = 200
	CodeIdenticalChains sdk.CodeType = 201
	CodeUnknownRequest  sdk.CodeType = sdk.CodeUnknownRequest
)

func codeToDefaultMsg(code sdk.CodeType) string {
	switch code {
	case CodeInvalidSequence:
		return "Invalid IBC packet sequence"
	case CodeIdenticalChains:
		return "Source and destination chain cannot be identical"
	default:
		return sdk.CodeToDefaultMsg(code)
	}
}

func ErrInvalidSequence(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeInvalidSequence, "")
}

func ErrIdenticalChains(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeIdenticalChains, "")
}

// -------------------------
// Helpers

func newError(codespace sdk.CodespaceType, code sdk.CodeType, msg string) sdk.Error {
	msg = msgOrDefaultMsg(msg, code)
	return sdk.NewError(codespace, code, msg)
}

func msgOrDefaultMsg(msg string, code sdk.CodeType) string {
	if msg != "" {
		return msg
	} else {
		return codeToDefaultMsg(code)
	}
}
