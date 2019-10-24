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
	CodeInvalidPacket              sdk.CodeType = 106
	CodeSequenceNotFound           sdk.CodeType = 107
	CodePacketTimeout              sdk.CodeType = 108
	CodeInvalidChannelState        sdk.CodeType = 109
)

// ErrChannelExists implements sdk.Error
func ErrChannelExists(codespace sdk.CodespaceType, channelID string) sdk.Error {
	return sdk.NewError(codespace, CodeChannelExists, fmt.Sprintf("channel with ID %s already exists", channelID))
}

// ErrChannelNotFound implements sdk.Error
func ErrChannelNotFound(codespace sdk.CodespaceType, channelID string) sdk.Error {
	return sdk.NewError(codespace, CodeChannelNotFound, fmt.Sprintf("channel with ID %s not found", channelID))
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

// ErrInvalidPacket implements sdk.Error
func ErrInvalidPacket(codespace sdk.CodespaceType, msg string) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidPacket, "invalid packet sequence counter")
}

// ErrSequenceNotFound implements sdk.Error
func ErrSequenceNotFound(codespace sdk.CodespaceType, seqType string) sdk.Error {
	return sdk.NewError(codespace, CodeSequenceNotFound, fmt.Sprintf("next %s sequence counter not found", seqType))
}

// ErrPacketTimeout implements sdk.Error
func ErrPacketTimeout(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodePacketTimeout, "packet timeout")
}

// ErrInvalidChannelState implements sdk.Error
func ErrInvalidChannelState(codespace sdk.CodespaceType, msg string) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidChannelState, msg)
}
