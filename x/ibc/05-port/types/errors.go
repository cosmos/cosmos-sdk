package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// connection error codes
const (
	DefaultCodespace sdk.CodespaceType = SubModuleName

	CodePortExists   sdk.CodeType = 101
	CodePortNotFound sdk.CodeType = 102
)

// ErrPortExists implements sdk.Error
func ErrPortExists(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodePortExists, "port already binded")
}

// ErrPortNotFound implements sdk.Error
func ErrPortNotFound(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodePortNotFound, "port not found")
}
