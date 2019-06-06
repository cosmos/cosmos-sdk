package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	DefaultCodespace sdk.CodespaceType = ModuleName

	CodeNoDenom               sdk.CodeType = 1
	CodeNoSigner              sdk.CodeType = 2
	CodeExchangeAlreadyExists              = 3
)

func ErrNoDenom(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeNoDenom, "denomination is empty")
}

func ErrNoSigner(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeNoSigner, "signer is nil")
}

func ErrExchangeAlreadyExists(codespace sdk.CodespaceType, msg string) sdk.Error {
	if msg != "" {
		return sdk.NewError(codespace, CodeExchangeAlreadyExists, msg)
	}
	return sdk.NewError(codespace, CodeExchangeAlreadyExists, "exchange already exists")
}
