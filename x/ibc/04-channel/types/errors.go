package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// channel error codes
const (
	DefaultCodespace sdk.CodespaceType = SubModuleName

	CodeChannelExists              sdk.CodeType = 219
	CodeChannelNotFound            sdk.CodeType = 220
	CodeInvalidCounterpartyChannel sdk.CodeType = 221
	CodeChannelCapabilityNotFound  sdk.CodeType = 222
	CodeInvalidPacket              sdk.CodeType = 223
	CodeSequenceNotFound           sdk.CodeType = 224
	CodePacketTimeout              sdk.CodeType = 225
	CodeInvalidChannel             sdk.CodeType = 226
	CodeInvalidChannelState        sdk.CodeType = 227
)

// ErrChannelExists implements sdk.Error
func ErrChannelExists(codespace sdk.CodespaceType, channelID string) error {
	return sdkerrors.New(
		string(codespace),
		uint32(CodeChannelExists),
		fmt.Sprintf("channel with ID %s already exists", channelID),
	)
}

// ErrChannelNotFound implements sdk.Error
func ErrChannelNotFound(codespace sdk.CodespaceType, portID, channelID string) error {
	return sdkerrors.New(
		string(codespace),
		uint32(CodeChannelNotFound),
		fmt.Sprintf("channel with ID %s on port %s not found", channelID, portID),
	)
}

// ErrInvalidCounterpartyChannel implements sdk.Error
func ErrInvalidCounterpartyChannel(codespace sdk.CodespaceType, msg string) error {
	return sdkerrors.New(
		string(codespace),
		uint32(CodeInvalidCounterpartyChannel),
		msg,
	)
}

// ErrChannelCapabilityNotFound implements sdk.Error
func ErrChannelCapabilityNotFound(codespace sdk.CodespaceType) error {
	return sdkerrors.New(
		string(codespace),
		uint32(CodeChannelCapabilityNotFound),
		"channel capability key not found",
	)
}

// ErrInvalidPacket implements sdk.Error
func ErrInvalidPacket(codespace sdk.CodespaceType, msg string) error {
	return sdkerrors.New(
		string(codespace),
		uint32(CodeInvalidPacket),
		msg,
	)
}

// ErrSequenceNotFound implements sdk.Error
func ErrSequenceNotFound(codespace sdk.CodespaceType, seqType string) error {
	return sdkerrors.New(
		string(codespace),
		uint32(CodeSequenceNotFound),
		fmt.Sprintf("next %s sequence counter not found", seqType),
	)
}

// ErrPacketTimeout implements sdk.Error
func ErrPacketTimeout(codespace sdk.CodespaceType) error {
	return sdkerrors.New(
		string(codespace),
		uint32(CodePacketTimeout),
		"packet timeout",
	)
}

// ErrInvalidChannel implements sdk.Error
func ErrInvalidChannel(codespace sdk.CodespaceType, msg string) error {
	return sdkerrors.New(
		string(codespace),
		uint32(CodeInvalidChannel),
		msg,
	)
}

// ErrInvalidChannelState implements sdk.Error
func ErrInvalidChannelState(codespace sdk.CodespaceType, msg string) error {
	return sdkerrors.New(
		string(codespace),
		uint32(CodeInvalidChannelState),
		msg,
	)
}
