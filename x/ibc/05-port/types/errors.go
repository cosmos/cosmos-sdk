package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// port error codes
const (
	DefaultCodespace sdk.CodespaceType = SubModuleName

	CodePortExists   sdk.CodeType = 228
	CodePortNotFound sdk.CodeType = 229
	CodeInvalidPort  sdk.CodeType = 230
)

// ErrPortExists implements sdk.Error
func ErrPortExists(codespace sdk.CodespaceType, portID string) error {
	return sdkerrors.New(
		string(codespace),
		uint32(CodePortExists),
		fmt.Sprintf("port with ID %s is already binded", portID),
	)
}

// ErrPortNotFound implements sdk.Error
func ErrPortNotFound(codespace sdk.CodespaceType, portID string) error {
	return sdkerrors.New(
		string(codespace),
		uint32(CodePortNotFound),
		fmt.Sprintf("port with ID %s not found", portID),
	)
}

// ErrInvalidPort implements sdk.Error
func ErrInvalidPort(codespace sdk.CodespaceType, msg string) error {
	return sdkerrors.New(
		string(codespace),
		uint32(CodePortNotFound),
		msg,
	)
}
