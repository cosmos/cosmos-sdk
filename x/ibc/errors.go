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

func ErrInvalidSequence() sdk.Error {
	return newError(CodeInvalidSequence, "")
}

func ErrIdenticalChains() sdk.Error {
	return newError(CodeIdenticalChains, "")
}

// -------------------------
// Helpers

func newError(code sdk.CodeType, msg string) sdk.Error {
	msg = msgOrDefaultMsg(msg, code)
	return sdk.NewError(code, msg)
}

func msgOrDefaultMsg(msg string, code sdk.CodeType) string {
	if msg != "" {
		return msg
	} else {
		return codeToDefaultMsg(code)
	}
}
