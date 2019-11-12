package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// connection error codes
const (
	DefaultCodespace sdk.CodespaceType = SubModuleName

	CodeConnectionExists              sdk.CodeType = 210
	CodeConnectionNotFound            sdk.CodeType = 211
	CodeClientConnectionPathsNotFound sdk.CodeType = 212
	CodeConnectionPath                sdk.CodeType = 213
	CodeInvalidCounterpartyConnection sdk.CodeType = 214
	CodeInvalidVersion                sdk.CodeType = 215
	CodeInvalidHeight                 sdk.CodeType = 216
	CodeInvalidConnectionState        sdk.CodeType = 217
	CodeInvalidCounterparty           sdk.CodeType = 218
)

// ErrConnectionExists implements sdk.Error
func ErrConnectionExists(codespace sdk.CodespaceType, connectionID string) error {
	return sdkerrors.New(
		string(codespace),
		uint32(CodeConnectionExists),
		fmt.Sprintf("connection with ID %s already exists", connectionID),
	)
}

// ErrConnectionNotFound implements sdk.Error
func ErrConnectionNotFound(codespace sdk.CodespaceType, connectionID string) error {
	return sdkerrors.New(
		string(codespace),
		uint32(CodeConnectionNotFound),
		fmt.Sprintf("connection with ID %s not found", connectionID),
	)
}

// ErrClientConnectionPathsNotFound implements sdk.Error
func ErrClientConnectionPathsNotFound(codespace sdk.CodespaceType, clientID string) error {
	return sdkerrors.New(
		string(codespace),
		uint32(CodeClientConnectionPathsNotFound),
		fmt.Sprintf("client connection paths not found for ID %s", clientID),
	)
}

// ErrConnectionPath implements sdk.Error
func ErrConnectionPath(codespace sdk.CodespaceType) error {
	return sdkerrors.New(
		string(codespace),
		uint32(CodeConnectionPath),
		"connection path is not associated to the client",
	)
}

// ErrInvalidCounterpartyConnection implements sdk.Error
func ErrInvalidCounterpartyConnection(codespace sdk.CodespaceType) error {
	return sdkerrors.New(
		string(codespace),
		uint32(CodeInvalidCounterpartyConnection),
		"couldn't verify connection membership on counterparty's client",
	)
}

// ErrInvalidHeight implements sdk.Error
func ErrInvalidHeight(codespace sdk.CodespaceType, msg string) error {
	return sdkerrors.New(
		string(codespace),
		uint32(CodeInvalidHeight),
		msg,
	)
}

// ErrInvalidConnectionState implements sdk.Error
func ErrInvalidConnectionState(codespace sdk.CodespaceType, msg string) error {
	return sdkerrors.New(
		string(codespace),
		uint32(CodeInvalidConnectionState),
		msg,
	)
}

// ErrInvalidCounterparty implements sdk.Error
func ErrInvalidCounterparty(codespace sdk.CodespaceType, msg string) error {
	return sdkerrors.New(
		string(codespace),
		uint32(CodeInvalidCounterparty),
		msg,
	)
}
