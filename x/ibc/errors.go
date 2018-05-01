package ibc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// IBC errors reserve 200 ~ 299.
const (
	DefaultCodespace sdk.CodespaceType = 3

	// IBC errors reserve 200 - 299.
	CodeInvalidSequence         sdk.CodeType = 200
	CodeIdenticalChains         sdk.CodeType = 201
	CodeChainMismatch           sdk.CodeType = 202
	CodeNoChannelOpened         sdk.CodeType = 203
	CodeChannelAlreadyOpened    sdk.CodeType = 204
	CodeUpdateCommitFailed      sdk.CodeType = 205
	CodeInvalidPacket           sdk.CodeType = 206
	CodeNoCommitFound           sdk.CodeType = 207
	CodeUnauthorizedSend        sdk.CodeType = 208
	CodeUnauthorizedSendReceipt sdk.CodeType = 209

	CodeUnknownRequest sdk.CodeType = sdk.CodeUnknownRequest
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

// nolint
func ErrInvalidSequence(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeInvalidSequence, "")
}
func ErrIdenticalChains(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeIdenticalChains, "")
}

func ErrChainMismatch(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeChainMismatch, "")
}

func ErrNoChannelOpened(codespace sdk.CodespaceType, srcChain string) sdk.Error {
	return newError(codespace, CodeNoChannelOpened, srcChain)
}

func ErrChannelAlreadyOpened(codespace sdk.CodespaceType, srcChain string) sdk.Error {
	return newError(codespace, CodeChannelAlreadyOpened, srcChain)
}

func ErrUpdateCommitFailed(codespace sdk.CodespaceType, err error) sdk.Error {
	return newError(codespace, CodeUpdateCommitFailed, err.Error())
}

func ErrInvalidPacket(codespace sdk.CodespaceType, err error) sdk.Error {
	return newError(codespace, CodeInvalidPacket, err.Error())
}

func ErrNoCommitFound(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeNoCommitFound, "")
}

func ErrUnauthorizedSend(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeUnauthorizedSend, "")
}

func ErrUnauthorizedSendReceipt(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeUnauthorizedSendReceipt, "")
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
	}
	return codeToDefaultMsg(code)
}
