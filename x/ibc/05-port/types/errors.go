package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// port error codes
const (
	DefaultCodespace sdk.CodespaceType = SubModuleName

	CodePortExists           sdk.CodeType = 101
	CodePortNotFound         sdk.CodeType = 102
	CodePortNotAuthenticated sdk.CodeType = 103
)

// ErrPortExists implements sdk.Error
func ErrPortExists(codespace sdk.CodespaceType, portID string) error {
	return sdkerrors.Register(
		string(codespace),
		uint32(CodePortExists),
		fmt.Sprintf("port with ID %s is already binded", portID),
	)
}

// ErrPortNotFound implements sdk.Error
func ErrPortNotFound(codespace sdk.CodespaceType, portID string) error {
	return sdkerrors.Register(
		string(codespace),
		uint32(CodePortNotFound),
		fmt.Sprintf("port with ID %s not found", portID),
	)
}

// ErrPortNotAuthenticated implements sdk.Error
func ErrPortNotAuthenticated(codespace sdk.CodespaceType, portID string) error {
	return sdkerrors.Register(
		string(codespace),
		uint32(CodePortNotAuthenticated),
		fmt.Sprintf("port with ID %s failed authentication", portID),
	)
}
