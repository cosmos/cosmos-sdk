package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// connection error codes
const (
	DefaultCodespace sdk.CodespaceType = SubModuleName

	CodeConnectionExists              sdk.CodeType = 101
	CodeConnectionNotFound            sdk.CodeType = 102
	CodeClientConnectionPathsNotFound sdk.CodeType = 103
	CodeConnectionPath                sdk.CodeType = 104
	CodeInvalidCounterpartyConnection sdk.CodeType = 105
	CodeInvalidVersion                sdk.CodeType = 106
	CodeInvalidHeight                 sdk.CodeType = 107
	CodeInvalidConnectionState        sdk.CodeType = 108
	CodeInvalidProof                  sdk.CodeType = 109
	CodeInvalidCounterparty           sdk.CodeType = 110
)

// ErrConnectionExists implements sdk.Error
func ErrConnectionExists(codespace sdk.CodespaceType, connectionID string) error {
	return sdkerrors.Register(
		string(codespace),
		uint32(CodeConnectionExists),
		fmt.Sprintf("connection with ID %s already exists", connectionID),
	)
}

// ErrConnectionNotFound implements sdk.Error
func ErrConnectionNotFound(codespace sdk.CodespaceType, connectionID string) error {
	return sdkerrors.Register(
		string(codespace),
		uint32(CodeConnectionNotFound),
		fmt.Sprintf("connection with ID %s not found", connectionID),
	)
}

// ErrClientConnectionPathsNotFound implements sdk.Error
func ErrClientConnectionPathsNotFound(codespace sdk.CodespaceType, clientID string) error {
	return sdkerrors.Register(
		string(codespace),
		uint32(CodeClientConnectionPathsNotFound),
		fmt.Sprintf("client connection paths not found for ID %s", clientID),
	)
}

// ErrConnectionPath implements sdk.Error
func ErrConnectionPath(codespace sdk.CodespaceType) error {
	return sdkerrors.Register(
		string(codespace),
		uint32(CodeConnectionPath),
		"connection path is not associated to the client",
	)
}

// ErrInvalidCounterpartyConnection implements sdk.Error
func ErrInvalidCounterpartyConnection(codespace sdk.CodespaceType) error {
	return sdkerrors.Register(
		string(codespace),
		uint32(CodeInvalidCounterpartyConnection),
		"couldn't verify connection membership on counterparty's client",
	)
}

// ErrInvalidHeight implements sdk.Error
func ErrInvalidHeight(codespace sdk.CodespaceType, msg string) error {
	return sdkerrors.Register(
		string(codespace),
		uint32(CodeInvalidHeight),
		msg,
	)
}

// ErrInvalidConnectionState implements sdk.Error
func ErrInvalidConnectionState(codespace sdk.CodespaceType, msg string) error {
	return sdkerrors.Register(
		string(codespace),
		uint32(CodeInvalidConnectionState),
		msg,
	)
}

// ErrInvalidConnectionProof implements sdk.Error
// TODO: move to ICS23
func ErrInvalidConnectionProof(codespace sdk.CodespaceType, msg string) error {
	return sdkerrors.Register(
		string(codespace),
		uint32(CodeInvalidProof),
		msg,
	)
}

// ErrInvalidCounterparty implements sdk.Error
func ErrInvalidCounterparty(codespace sdk.CodespaceType, msg string) error {
	return sdkerrors.Register(
		string(codespace),
		uint32(CodeInvalidCounterparty),
		msg,
	)
}
