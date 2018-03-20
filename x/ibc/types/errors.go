package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// IBC errors reserve 200 - 299.
	CodeInvalidSequence sdk.CodeType = 200
	CodeIdenticalChains sdk.CodeType = 201
	CodeChainMismatch   sdk.CodeType = 202
	CodeUnknownRequest  sdk.CodeType = sdk.CodeUnknownRequest
)

func codeToDefaultMsg(code sdk.CodeType) string {
	switch code {
	case CodeInvalidSequence:
		return "Invalid IBC packet sequence"
	case CodeIdenticalChains:
		return "Source and destination chain cannot be identical"
	case CodeChainMismatch:
		return "Destination chain is not current chain"
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

func ErrChainMismatch() sdk.Error {
	return newError(CodeChainMismatch, "")
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
