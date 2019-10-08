package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// connection error codes
const (
	DefaultCodespace sdk.CodespaceType = SubModuleName

	CodeConnectionExists              sdk.CodeType = 101
	CodeConnectionNotFound            sdk.CodeType = 102
	CodeClientConnectionPathsNotFound sdk.CodeType = 103
	CodeConnectionPath                sdk.CodeType = 104
)

// ErrConnectionExists implements sdk.Error
func ErrConnectionExists(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeConnectionExists, "connection already exists")
}

// ErrConnectionNotFound implements sdk.Error
func ErrConnectionNotFound(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeConnectionNotFound, "connection not found")
}

// ErrClientConnectionPathsNotFound implements sdk.Error
func ErrClientConnectionPathsNotFound(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeClientConnectionPathsNotFound, "client connection paths not found")
}

// ErrConnectionPath implements sdk.Error
func ErrConnectionPath(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeConnectionPath, "connection path is not associated to the client")
}
