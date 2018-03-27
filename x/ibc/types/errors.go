package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// IBC errors reserve 200 - 299.
	CodeInvalidSequence      sdk.CodeType = 200
	CodeIdenticalChains      sdk.CodeType = 201
	CodeChainMismatch        sdk.CodeType = 202
	CodeUnknownRequest       sdk.CodeType = sdk.CodeUnknownRequest
	CodeNoChannelOpened      sdk.CodeType = 203
	CodeChannelAlreadyOpened sdk.CodeType = 204
	CodeUpdateCommitFailed   sdk.CodeType = 205
	CodeInvalidPacket        sdk.CodeType = 206
	CodeNoCommitFound        sdk.CodeType = 207
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

func ErrNoChannelOpened(srcChain string) sdk.Error {
	return newError(CodeNoChannelOpened, srcChain)
}

func ErrChannelAlreadyOpened(srcChain string) sdk.Error {
	return newError(CodeChannelAlreadyOpened, srcChain)
}

func ErrUpdateCommitFailed(err error) sdk.Error {
	return newError(CodeUpdateCommitFailed, err.Error())
}

func ErrInvalidPacket(err error) sdk.Error {
	return newError(CodeInvalidPacket, err.Error())
}

func ErrNoCommitFound() sdk.Error {
	return newError(CodeNoCommitFound, "")
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
