package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// transfer error codes
const (
	DefaultCodespace sdk.CodespaceType = SubModuleName

	CodeInvalidAddress      sdk.CodeType = 101
	CodeErrSendPacket       sdk.CodeType = 102
	CodeInvalidPacketData   sdk.CodeType = 103
	CodeInvalidChannelOrder sdk.CodeType = 104
	CodeInvalidPort         sdk.CodeType = 105
	CodeInvalidVersion      sdk.CodeType = 106
	CodeProofMissing        sdk.CodeType = 107
	CodeInvalidHeight       sdk.CodeType = 108
)

// ErrInvalidAddress implements sdk.Error
func ErrInvalidAddress(codespace sdk.CodespaceType, msg string) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidAddress, msg)
}

// ErrSendPacket implements sdk.Error
func ErrSendPacket(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeErrSendPacket, "failed to send packet")
}

// ErrInvalidPacketData implements sdk.Error
func ErrInvalidPacketData(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidPacketData, "invalid packet data")
}

// ErrInvalidChannelOrder implements sdk.Error
func ErrInvalidChannelOrder(codespace sdk.CodespaceType, order string) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidChannelOrder, fmt.Sprintf("invalid channel order: %s", order))
}

// ErrInvalidPort implements sdk.Error
func ErrInvalidPort(codespace sdk.CodespaceType, portID string) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidPort, fmt.Sprintf("invalid port ID: %s", portID))
}

// ErrInvalidVersion implements sdk.Error
func ErrInvalidVersion(codespace sdk.CodespaceType, msg string) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidVersion, msg)
}
