package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CodeType definition
type CodeType = sdk.CodeType

// Supply errors reserve 550 - 599
const (
	DefaultCodespace sdk.CodespaceType = "supply"

	CodeInvalidTokenHolder CodeType = 550
	CodeUnknownTokenHolder CodeType = 551
)

// ErrInvalidTokenHolder is an error
func ErrInvalidTokenHolder(codespace sdk.CodespaceType, msg string) sdk.Error {
	if msg != "" {
		return sdk.NewError(codespace, CodeInvalidTokenHolder, msg)
	}
	return sdk.NewError(codespace, CodeInvalidTokenHolder, "invalid token holder")
}

// ErrUnknownTokenHolder is an error
func ErrUnknownTokenHolder(codespace sdk.CodespaceType, msg string) sdk.Error {
	if msg != "" {
		return sdk.NewError(codespace, CodeUnknownTokenHolder, msg)
	}
	return sdk.NewError(codespace, CodeUnknownTokenHolder, "unknown token holder")
}
