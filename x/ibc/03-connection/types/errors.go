package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// client error codes
const (
	DefaultCodespace sdk.CodespaceType = SubModuleName

	CodeClientExists     sdk.CodeType = 101
	CodeClientNotFound   sdk.CodeType = 102
	CodeClientFrozen     sdk.CodeType = 103
	CodeInvalidConsensus sdk.CodeType = 104
	CodeValidatorJailed  sdk.CodeType = 104
)

// ErrClientExists implements sdk.Error
func ErrClientExists(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeClientExists, "client already exists")
}

// ErrClientNotFound implements sdk.Error
func ErrClientNotFound(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeClientNotFound, "client not found")
}

// ErrClientFrozen implements sdk.Error
func ErrClientFrozen(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeClientFrozen, "client is frozen due to misbehaviour")
}

// ErrInvalidConsensus implements sdk.Error
func ErrInvalidConsensus(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidConsensus, "invalid consensus algorithm type")
}
