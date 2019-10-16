package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// port error codes
const (
	DefaultCodespace sdk.CodespaceType = SubModuleName

	CodePortExists           sdk.CodeType = 101
	CodePortNotFound         sdk.CodeType = 102
	CodePortNotAuthenticated sdk.CodeType = 103
	CodeInvalidPortID        sdk.CodeType = 104
)

// ErrPortExists implements sdk.Error
func ErrPortExists(codespace sdk.CodespaceType, portID string) sdk.Error {
	return sdk.NewError(codespace, CodePortExists, fmt.Sprintf("port with ID %s is already binded", portID))
}

// ErrPortNotFound implements sdk.Error
func ErrPortNotFound(codespace sdk.CodespaceType, portID string) sdk.Error {
	return sdk.NewError(codespace, CodePortNotFound, fmt.Sprintf("port with ID %s not found", portID))
}

// ErrPortNotAuthenticated implements sdk.Error
func ErrPortNotAuthenticated(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodePortNotAuthenticated, "port failed authentication")
}

// ErrInvalidPortID implements sdk.Error
func ErrInvalidPortID(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidPortID, "invalid port ID")
}
