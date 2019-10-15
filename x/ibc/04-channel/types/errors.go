package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// channel error codes
const (
	DefaultCodespace sdk.CodespaceType = SubModuleName

	CodeChannelExists              sdk.CodeType = 101
	CodeChannelNotFound            sdk.CodeType = 102
	CodeInvalidConnectionHops      sdk.CodeType = 103
	CodeInvalidCounterpartyChannel sdk.CodeType = 104
	CodeChannelCapabilityNotFound  sdk.CodeType = 105
	CodeInvalidPacketSequence      sdk.CodeType = 106
	CodeSequenceNotFound           sdk.CodeType = 107
	CodePacketTimeout              sdk.CodeType = 108
	CodeChanIDLen                  sdk.CodeType = 109
	CodePortIDLen                  sdk.CodeType = 110
)

// ErrChannelExists implements sdk.Error
func ErrChannelExists(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeChannelExists, "channel already exists")
}

// ErrChannelNotFound implements sdk.Error
func ErrChannelNotFound(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeChannelNotFound, "channel not found")
}

// ErrInvalidConnectionHops implements sdk.Error
func ErrInvalidConnectionHops(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidConnectionHops, "IBC v1 only supports one connection hop")
}

// ErrInvalidCounterpartyChannel implements sdk.Error
func ErrInvalidCounterpartyChannel(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidCounterpartyChannel, "counterparty channel doesn't match the expected one")
}

// ErrChannelCapabilityNotFound implements sdk.Error
func ErrChannelCapabilityNotFound(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeChannelCapabilityNotFound, "channel capability key not found")
}

// ErrInvalidPacketSequence implements sdk.Error
func ErrInvalidPacketSequence(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidPacketSequence, "invalid packet sequence counter")
}

// ErrSequenceNotFound implements sdk.Error
func ErrSequenceNotFound(codespace sdk.CodespaceType, seqType string) sdk.Error {
	return sdk.NewError(codespace, CodeSequenceNotFound, fmt.Sprintf("next %s sequence counter not found", seqType))
}

// ErrPacketTimeout implements sdk.Error
func ErrPacketTimeout(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodePacketTimeout, "packet timeout")
}

// ErrChanIDLen implements sdk.Error
func ErrLenChanID(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeChanIDLen, "chanid too long")
}

// ErrLenPortID implements sdk.Error
func ErrLenPortID(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodePortIDLen, "portid too long")
}
